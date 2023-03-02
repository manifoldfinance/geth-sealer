{
  description = "Geth on Nix";
  nixConfig = {
    extra-substituters = [
      "https://nix-community.cachix.org"
    ];
    extra-trusted-public-keys = [
      "nix-community.cachix.org-1:mB9FSh9qf2dCimDSUo8Zy7bkq5CX+/rkCWyvRCYg3Fs="
    ];
  };
  inputs = {
    # packages
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    ethereum-nix = {
      url = "github:nix-community/ethereum.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    # flake-parts
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    flake-root.url = "github:srid/flake-root";
    mission-control.url = "github:Platonic-Systems/mission-control";
  };
  outputs = inputs @ {
    flake-parts,
    flake-root,
    mission-control,
    nixpkgs,
    ...
  }:
    (flake-parts.lib.evalFlakeModule
      {inherit inputs;}
      {
        imports = [
          flake-root.flakeModule
          mission-control.flakeModule
        ];
        systems = ["x86_64-linux"];
        perSystem = {
          self',
          config,
          pkgs,
          lib,
          ...
        }: let
          inherit (config.mission-control) installToDevShell;
          inherit (pkgs) mkShellNoCC;
          inherit (lib) makeLibraryPath;
        in {
          devShells.default = installToDevShell (mkShellNoCC {
            name = "geth";
            packages = with pkgs; [
              go
              httpie
              websocat
            ];
          });
        };
      })
    .config
    .flake;
}
