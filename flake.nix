{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };
  
  outputs = {nixpkgs, ...}: let
    lib = nixpkgs.lib;
    forAllSystems = f: lib.genAttrs lib.systems.flakeExposed (system: f {
      pkgs = nixpkgs.legacyPackages.${system};
    });
  in {
    devShells = forAllSystems ({pkgs, ...}: {
      default = pkgs.mkShell {
        packages = with pkgs; [
          go
          gopls
          wezterm
          ghostty
        ];
      };
    });
  };
}
