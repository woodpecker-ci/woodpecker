{
  # Override nixpkgs to use the latest set of node packages
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # generic
            gnumake
            gnutar

            # frontend
            nodejs_20
            nodePackages.pnpm
            nodePackages.typescript
            nodePackages.typescript-language-server

            # backend
            go_1_22
            gofumpt
            golangci-lint
            go-mockery
            protobuf
          ];
        };
      }
    );
}
