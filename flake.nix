{
  description = "Basic rust flake :)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    rust-overlay.url = "github:oxalica/rust-overlay";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      #self,
      nixpkgs,
      rust-overlay,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        overlays = [ (import rust-overlay) ];
        pkgs = import nixpkgs {
          inherit system overlays;
        };

        rustToolchain = pkgs.rust-bin.stable.latest.default.override {
          extensions = [ "rust-src" ];
          targets = [ "wasm32-unknown-unknown" ];
        };
      in
      with pkgs;
      {
        devShells.default = mkShell {
          LD_LIBRARY_PATH = lib.makeLibraryPath [ openssl ];
          buildInputs = [
            openssl
            pkg-config
            eza
            fd
            rustToolchain
            rust-analyzer
            pkgs.zsh
            cmake
            ## tauri
            librsvg
            webkitgtk_4_1
            ## tauri end
            knot-dns # kdig
            dnsperf # per testing
            air
            goose
            sqlite
          ];
          shellHook = ''
            alias ls=eza
            export PATH=$PATH:${pkgs.rust-analyzer}/bin
            alias find=fd
          '';
        };
      }
    );
}
