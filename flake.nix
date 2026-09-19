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
            go = go_1_27;

            # rebuild Go based tools with the go version above
            buildGoModule = pkgs.buildGoModule.override { inherit go; };
            withGo =
              pkg:
              let
                builderArgs = lib.filterAttrs (name: _: lib.hasPrefix "buildGo" name) (
                  if pkg ? override then lib.functionArgs pkg.override else { }
                );
              in
              if builderArgs == { } then pkg else pkg.override (lib.mapAttrs (_: _: buildGoModule) builderArgs);
          in
          pkgs.mkShell {
            buildInputs = map withGo [
              # generic
              gnumake
              gnutar
              gzip
              zip
              tree

              # frontend
              nodejs_24
              pnpm
              typescript
              typescript-language-server

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

              # docs
              graphviz
            ];
            CFLAGS = "-I${pkgs.glibc.dev}/include";
            LDFLAGS = "-L${pkgs.glibc}/lib";
            GO = "${go}/bin/go";
            GOROOT = "${go}/share/go";
          };
      }
    );
}
