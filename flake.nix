{
  description = "wkey: A Voice Input Utility with Wayland support";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:
    utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        
        # UI and System dependencies
        runtimeDeps = with pkgs; [
          wl-clipboard
          wtype
        ];

        buildDeps = with pkgs; [
          pkg-config
          libGL
          xorg.libX11
          xorg.libXcursor
          xorg.libXrandr
          xorg.libXinerama
          xorg.libXi
          xorg.libXxf86vm
          libxkbcommon
        ];
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "wkey";
          version = "0.1.0";
          src = ./.;

          # Run `nix build` and replace this with the hash provided in the error message
          vendorHash = "sha256-STCJLN9oawV4/zNR9iYTjcdt9i9fXrl8gdiTUoQlJzY=";

          nativeBuildInputs = [ pkgs.pkg-config pkgs.makeWrapper ];
          buildInputs = buildDeps;

          subPackages = [ "cmd/wkey" ];

          postInstall = ''
            wrapProgram $out/bin/wkey \
              --prefix PATH : ${pkgs.lib.makeBinPath runtimeDeps}
          '';
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            go-tools
          ] ++ buildDeps ++ runtimeDeps;

          LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath buildDeps;
        };
      }
    );
}
