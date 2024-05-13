{
  # Override nixpkgs to use the latest set of node packages
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/master";
  inputs.systems.url = "github:nix-systems/default";

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      systems,
    }:
    flake-utils.lib.eachSystem (import systems) (
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
