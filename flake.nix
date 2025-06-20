{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=master";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        # Common attributes for all components
        commonAttrs = {
          version = "0.0.0";
          vendorHash = "sha256-0hInlX2yXf9IBW1h9lYeG1pn9v3LtVRoWNJJIiIdPaU=";
        };

        # Common builder function for woodpecker components
        mkWoodpeckerComponent =
          {
            pname,
            subPackages ? [ ],
            buildTags ? [ ],
            nativeBuildInputs ? [ ],
            preBuild ? "",
            CGO_ENABLED ? "0",
          }:
          pkgs.buildGoModule ({
            inherit pname;
            inherit (commonAttrs)
              version
              vendorHash
              ;
            inherit
              subPackages
              buildTags
              nativeBuildInputs
              preBuild
              CGO_ENABLED
              ;

            src = ./.;

            meta = {
              mainProgram = pname;
              description = "A distributed CI/CD system";
              homepage = "https://woodpecker-ci.org";
              license = pkgs.lib.licenses.asl20;
            };
          });

      in
      {
        packages = rec {
          cli = mkWoodpeckerComponent {
            pname = "woodpecker-cli";
            subPackages = [ "cmd/cli" ];
          };

          server = mkWoodpeckerComponent {
            pname = "woodpecker-server";
            subPackages = [ "cmd/server" ];
            CGO_ENABLED = "1";
            nativeBuildInputs = with pkgs; [
              nodejs_20
              pnpm
              gnumake
            ];
            preBuild = ''
              export HOME=$(mktemp -d)

              cd web
              pnpm install --frozen-lockfile
              pnpm build
              cd ..

              go generate cmd/server/swagger.go
            '';
          };

          agent = mkWoodpeckerComponent {
            pname = "woodpecker-agent";
            subPackages = [ "cmd/agent" ];
          };

          default = pkgs.symlinkJoin {
            name = "woodpecker";
            paths = [
              cli
              agent
              # server # need impure because of pnpm
            ];
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # generic
            gnumake
            gnutar
            zip
            tree
            git

            # frontend
            nodejs_24
            pnpm
            nodePackages.typescript
            nodePackages.typescript-language-server

            # backend
            go_1_24
            glibc.static
            gofumpt
            golangci-lint
            go-mockery
            protobuf
            sqlite
            go-swag # for generate-openapi
            addlicense
            protoc-gen-go
            protoc-gen-go-grpc
            gcc
          ];
          CFLAGS = "-I${pkgs.glibc.dev}/include";
          LDFLAGS = "-L${pkgs.glibc}/lib";
        };
      }
    );
}
