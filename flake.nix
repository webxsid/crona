{
  description = "Crona, a local-first work tracker for the terminal";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      inherit (nixpkgs) lib;
      # Intel macOS is left out because nixpkgs no longer supports it.
      systems = [ "aarch64-darwin" "aarch64-linux" "x86_64-linux" ];
      forAllSystems = lib.genAttrs systems;

      # Each channel's release, rewritten by scripts/publish_nix_sources.sh on each release. A
      # release with a rev is built from source at that tag; older ones, from before Crona was
      # one Go module, still carry per-platform hashes of the release bundles.
      sources = builtins.fromJSON (builtins.readFile ./nix/sources.json);
      # The vendored modules' hash for this tree; scripts/update_nix_vendor_hash.sh refreshes it.
      vendor = builtins.fromJSON (builtins.readFile ./nix/vendor.json);

      # The GitHub repository releases are built from; keep in step with RepoOwner.
      owner = "crona-app";
      repo = "crona";

      # The version this tree declares, for builds of the tree itself.
      treeVersion =
        let
          lines = lib.splitString "\n" (builtins.readFile ./shared/version/version.go);
          found = lib.findFirst (line: builtins.match ''var Version = "([^"]+)"'' line != null) null lines;
        in
        builtins.head (builtins.match ''var Version = "([^"]+)"'' found);

      meta = provenance: {
        description = "Local-first work tracker for the terminal";
        homepage = "https://crona.work";
        license = lib.licenses.mit;
        mainProgram = "crona";
        platforms = systems;
        sourceProvenance = [ provenance ];
      };

      # The crona CLI, crona-daemon and crona-tui, built from source as the release builds them.
      buildCrona = pkgs: { pname, version, src, vendorHash }:
        pkgs.buildGoModule {
          inherit pname version src vendorHash;
          subPackages = [ "cli/cmd/crona" "kernel/cmd/crona-daemon" "tui/cmd/crona-tui" ];
          env.CGO_ENABLED = 0;
          ldflags = [ "-s" "-w" "-X crona.work/shared/version.Version=${version}" ];
          # The tests need a writable home and local sockets; CI runs them.
          doCheck = false;
          meta = meta lib.sourceTypes.fromSource;
        };

      # Releases from before the single module ship as static Go binaries, unpacked as they are.
      goarch = { aarch64-darwin = "darwin-arm64"; aarch64-linux = "linux-arm64"; x86_64-linux = "linux-amd64"; };
      unpackBundle = pkgs: pname: source:
        pkgs.stdenvNoCC.mkDerivation {
          inherit pname;
          version = source.version;
          src = pkgs.fetchurl {
            url = "https://github.com/${owner}/${repo}/releases/download/v${source.version}/crona-bundle-${source.version}-${goarch.${pkgs.stdenv.hostPlatform.system}}.zip";
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
          meta = meta lib.sourceTypes.binaryNativeCode;
        };

      release = pkgs: pname: source:
        if source ? rev then
          buildCrona pkgs {
            inherit pname;
            inherit (source) version vendorHash;
            src = pkgs.fetchFromGitHub { inherit owner repo; inherit (source) rev hash; };
          }
        else
          unpackBundle pkgs pname source;
    in
    {
      packages = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in lib.mapAttrs (release pkgs) sources // {
          # This checkout, built from source: nix build github:crona-app/crona#crona-git.
          crona-git = buildCrona pkgs {
            pname = "crona-git";
            version = treeVersion;
            src = lib.fileset.toSource {
              root = ./.;
              fileset = lib.fileset.unions [ ./go.mod ./go.sum ./assets ./cli ./kernel ./shared ./tui ];
            };
            inherit (vendor) vendorHash;
          };
          default = self.packages.${system}.crona;
        });
    };
}
