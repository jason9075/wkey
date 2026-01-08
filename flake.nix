{
  description = "A very basic flake";

  inputs = { nixpkgs.url = "github:nixos/nixpkgs/24.05"; };

  outputs = { self, nixpkgs }:
    let
      pkgs-x86 = import nixpkgs { system = "x86_64-linux"; };
      pkgs-mac = import nixpkgs { system = "x86_64-darwin"; };
    in {
      devShells = {
        x86_64-linux.default = pkgs-x86.mkShell {
          nativeBuildInputs = with pkgs-x86; [ go entr ];

          shellHook = ''
            echo "Environment activated."
          '';
        };

        x86_64-darwin.default = pkgs-mac.mkShell {
          nativeBuildInputs = with pkgs-mac; [ go ];

          shellHook = ''
            echo "Environment activated."
          '';
        };
      };
    };
}
