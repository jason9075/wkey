{
  description = "Voice Input Utility Development Environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" ];
      forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs { inherit system; };
      });
    in
    {
      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            # Go (default version)
            go
            gopls
            gotools
            go-tools

            # Audio
            pipewire
            wireplumber

            # Wayland / Clipboard
            wl-clipboard
            wtype
            
            # UI deps (if using Fyne or GTK)
            pkg-config
            libGL
            xorg.libX11
            xorg.libXcursor
            xorg.libXrandr
            xorg.libXinerama
            xorg.libXi
            xorg.libXxf86vm
          ];
        };
      });
    };
}
