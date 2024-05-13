{
  # Override nixpkgs to use the latest set of node packages
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/master";
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
            nodejs
            nodePackages.pnpm
            nodePackages.typescript
            nodePackages.typescript-language-server

            # backend
            go
            gofumpt
            golangci-lint
            go-mockery
            protobuf
          ];
        };
      }
    );
}
