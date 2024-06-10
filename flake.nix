{
  description = "Generate RSS feed for GH issues and PRs";

  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        deploy = pkgs.writeShellScriptBin "deploy" ''
          set -euo pipefail
          ${pkgs.flyctl}/bin/fly deploy
        '';
      in
      {
        packages = rec {
          noah = pkgs.buildGoModule {
            pname = "gh-issues-to-rss";
            version = "dev";
            src = ./.;
            vendorHash = "sha256-30tnt3fYXDDm+YQgY65dUcytDokrUdCChOU5LMgnGYE=";
            doCheck = false;
          };

          default = noah;
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go

            # for deployment
            flyctl
            deploy
          ];
        };
      }
    );
}