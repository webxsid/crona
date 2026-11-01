{
  description = "Crona, a local-first work tracker for the terminal";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      # Intel macOS is left out because nixpkgs no longer supports it.
      # Release hashes per channel, rewritten by scripts/publish_nix_sources.sh on each release.
      sources = builtins.fromJSON (builtins.readFile ./nix/sources.json);
      goarch = { aarch64-darwin = "darwin-arm64"; aarch64-linux = "linux-arm64"; x86_64-linux = "linux-amd64"; };
      forAllSystems = nixpkgs.lib.genAttrs (builtins.attrNames goarch);

      # The release bundles are static Go binaries, so they are unpacked as they are.
      mkCrona = pkgs: pname: source:
        pkgs.stdenvNoCC.mkDerivation {
          inherit pname;
          version = source.version;
          src = pkgs.fetchurl {
            url = "https://github.com/webxsid/crona/releases/download/v${source.version}/crona-bundle-${source.version}-${goarch.${pkgs.stdenv.hostPlatform.system}}.zip";
            sha256 = source.platforms.${pkgs.stdenv.hostPlatform.system};
          };
          nativeBuildInputs = [ pkgs.unzip ];
          sourceRoot = ".";
          dontConfigure = true;
          dontBuild = true;
          dontStrip = true;
          installPhase = ''
            runHook preInstall
            install -Dm755 crona crona-daemon crona-tui -t $out/bin
            runHook postInstall
          '';
          meta = {
            description = "Local-first work tracker for the terminal";
            homepage = "https://crona.work";
            license = pkgs.lib.licenses.mit;
            mainProgram = "crona";
            platforms = builtins.attrNames goarch;
            sourceProvenance = [ pkgs.lib.sourceTypes.binaryNativeCode ];
          };
        };
    in
    {
      packages = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in nixpkgs.lib.mapAttrs (mkCrona pkgs) sources // { default = self.packages.${system}.crona; });
    };
}
