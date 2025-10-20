{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=master";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default =
          with pkgs;
          let
            go = go_1_24;
          in
          pkgs.mkShell {
            buildInputs = [
              # generic
              gnumake
              gnutar
              gzip
              zip
              tree

              # frontend
              nodejs_24
              pnpm
              nodePackages.typescript
              nodePackages.typescript-language-server

              # backend
              go
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
            GO = "${go}/bin/go";
            GOROOT = "${go}/share/go";
          };
      }
    );
}
