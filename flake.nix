{
  description = "Timeful development shell";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/79b35bf0bda5cd110f856aa5b5b2c5ba4460dbf5";
    systems.url = "github:nix-systems/default/future-26.11";
    flake-parts = {
      url = "github:hercules-ci/flake-parts/17c9d6cdfc60c64f4ee8d306f9bc0b4ccb51481e";
      inputs.nixpkgs-lib.url = "github:nix-community/nixpkgs.lib";
    };
    backlog-md.url = "github:time-tools/Backlog.md/aded8e254e6a0205b878cf07e631d1a592782040";
  };

  outputs = inputs@{ flake-parts, systems, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import systems;

      perSystem = { pkgs, system, ... }:
        let
          e2e = pkgs.writeShellScriptBin "e2e" ''
            set -euo pipefail
            export PATH="${pkgs.nodejs_26}/bin:$PATH"
            export PLAYWRIGHT_BROWSERS_PATH="${pkgs.playwright-driver.browsers}"
            REPO_ROOT="$(git rev-parse --show-toplevel)"
            cd "$REPO_ROOT/frontend"
            npm ci
            cd "$REPO_ROOT/e2e"
            npm ci
            exec npm run test:e2e -- "$@"
          '';
        in {
          packages = { inherit e2e; };
          apps.e2e = {
            type = "app";
            program = "${e2e}/bin/e2e";
          };
          devShells.default = pkgs.mkShell {
          packages = [
            pkgs.nodejs_26
            pkgs.python3
            pkgs.go
            pkgs.playwright-driver.browsers
            inputs.backlog-md.packages.${system}.default
            pkgs.codebase-memory-mcp
            pkgs.ripgrep
            pkgs.actionlint
          ];
          shellHook = ''
             export BACKLOG_CWD="$PWD"
             export PLAYWRIGHT_BROWSERS_PATH="${pkgs.playwright-driver.browsers}"
          '';
        };
      };
    };
}
