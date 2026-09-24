import { execFileSync } from "node:child_process"
import { createHmac, randomBytes } from "node:crypto"
import { fileURLToPath } from "node:url"

const repositoryRoot = fileURLToPath(new URL("../../", import.meta.url))

const composeArguments = [
  "compose",
  "--env-file",
  ".env.test",
  "-f",
  "compose.yaml",
  "-f",
  "compose.test.yaml",
]

let isolatedDatabaseName: string | undefined

// Resolves the Playwright-owned PostgreSQL database from the application URI
// the isolated server-test container actually connected to, so inspections can
// never read a development database.
function isolatedDatabase(): string {
  if (isolatedDatabaseName) {
    return isolatedDatabaseName
  }

  const uri = execFileSync(
    "docker",
    [
      ...composeArguments,
      "exec",
      "-T",
      "server-test",
      "printenv",
      "POSTGRES_APPLICATION_URI",
    ],
    { cwd: repositoryRoot, encoding: "utf8" },
  ).trim()
  const database = new URL(uri).pathname.slice(1)
  if (!database.startsWith("timeful-test-")) {
    throw new Error(
      "Database inspection requires Playwright's isolated test database",
    )
  }
  isolatedDatabaseName = database
  return database
}

// Runs a scalar SQL statement against the isolated database and returns the
// unaligned, tuple-only result, so callers can assert on exact values.
export function databaseScalar(sql: string): string {
  const database = isolatedDatabase()
  return execFileSync(
    "docker",
    [
      ...composeArguments,
      "exec",
      "-T",
      "postgres-test",
      "sh",
      "-ec",
      'psql --username "$POSTGRES_USER" --dbname "$1" --set=ON_ERROR_STOP=1 --tuples-only --no-align --command "$2"',
      "database-scalar",
      database,
      sql,
    ],
    { cwd: repositoryRoot, encoding: "utf8" },
  ).trim()
}

// Runs a non-returning SQL statement against the isolated database.
export function databaseExec(sql: string): void {
  const database = isolatedDatabase()
  execFileSync(
    "docker",
    [
      ...composeArguments,
      "exec",
      "-T",
      "postgres-test",
      "sh",
      "-ec",
      'psql --username "$POSTGRES_USER" --dbname "$1" --set=ON_ERROR_STOP=1 --command "$2"',
      "database-exec",
      database,
      sql,
    ],
    { cwd: repositoryRoot, encoding: "utf8" },
  )
}

// Stores the only active OTP challenge for an email using the same salted,
// one-way hash the server writes, so sign-in journeys can seed a challenge
// without ever reading the plaintext code back from the database.
export function seedOtpChallenge(email: string, code: string): void {
  const salt = randomBytes(16)
  const digest = createHmac("sha256", salt).update(code).digest()
  const codeHash = `${salt.toString("base64")}:${digest.toString("base64")}`
  const literal = (value: string) => `'${value.replace(/'/g, "''")}'`
  databaseExec(
    `INSERT INTO otp_challenges (email, code_hash, expires_at, attempts) VALUES (${literal(
      email.trim().toLowerCase(),
    )}, ${literal(codeHash)}, clock_timestamp() + interval '10 minutes', 0) ` +
      `ON CONFLICT (email) DO UPDATE SET code_hash = EXCLUDED.code_hash, expires_at = EXCLUDED.expires_at, attempts = 0, updated_at = clock_timestamp()`,
  )
}
