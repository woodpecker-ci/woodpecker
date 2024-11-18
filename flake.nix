{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
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
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # generic
            gnumake
            gnutar
            zip
            tree

            # frontend
            nodejs_20
            pnpm
            nodePackages.typescript
            nodePackages.typescript-language-server

            # backend
            go_1_22
            glibc.static
            gofumpt
            golangci-lint
            go-mockery
            protobuf
            sqlite
            go-swag # for generate-swagger
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
