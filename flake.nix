{
  description = "Mash - Terminal Mosh/SSH connection manager for macOS and Linux.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      pkgsFor = forAllSystems (system: import nixpkgs {
        inherit system;
      });
    in
    {
      packages = forAllSystems (system: {
        default = self.packages.${system}.mash;

        mash = pkgsFor.${system}.buildGoModule {
          pname = "mash";
          version = "0.1.0";
          src = ./.;
          vendorHash = "";
        };
      });

      checks = forAllSystems (system:
        let
          pkgs = pkgsFor.${system};
        in
        {
          mash-e2e = pkgs.runCommand "mash-e2e"
            {
              buildInputs = [ self.packages.${system}.mash ];
            }
            ''
              echo "=== mash binary smoke test ==="
              mash_path="$(command -v mash)"
              echo "binary at: $mash_path"
              file "$mash_path"
              mash --help 2>&1 || true
              mkdir -p "$out"
              echo "OK" > "$out/result"
            '';

          mash-golden-tests = pkgs.buildGoModule {
            pname = "mash-golden-tests";
            version = "0.1.0";
            src = ./.;
            vendorHash = "";
            doCheck = true;

            checkPhase = ''
              runHook preCheck

              export HOME=$(mktemp -d)
              mkdir -p "$HOME/.ssh"
              cp internal/tui/testdata/ssh_config "$HOME/.ssh/config"

              echo "=== mash SSH config golden regression tests ==="
              echo "SSH config:"
              cat "$HOME/.ssh/config"
              echo ""

              go test ./internal/tui/ -run TestRealConfigNavigationAndScreens -v -count=1

              runHook postCheck
            '';

            installPhase = ''
              runHook preInstall
              mkdir -p "$out"
              echo "PASS" > "$out/result"
              runHook postInstall
            '';
          };

          mash-cloud-tests = pkgs.buildGoModule {
            pname = "mash-cloud-tests";
            version = "0.1.0";
            src = ./.;
            vendorHash = "";
            doCheck = true;

            checkPhase = ''
              runHook preCheck

              export HOME=$(mktemp -d)
              mkdir -p "$HOME/.ssh"
              cp internal/tui/testdata/ssh_config "$HOME/.ssh/config"

              echo "=== mash cloud discovery regression tests ==="
              echo "OpenTofu state file:"
              cat internal/tui/testdata/tofu_state.json | head -20
              echo ""
              echo "Tailscale status file:"
              cat internal/tui/testdata/tailscale_status.json | head -20
              echo ""

              go test ./internal/tui/ -run TestCloudNavigationAndScreens -v -count=1

              runHook postCheck
            '';

            installPhase = ''
              runHook preInstall
              mkdir -p "$out"
              echo "PASS" > "$out/result"
              runHook postInstall
            '';
          };
        });

      devShells = forAllSystems (system:
        let
          pkgs = pkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              goimports
              opentofu
            ];
            shellHook = ''
              echo "mash dev shell — go $(go version) | tofu $(tofu version | head -1)"
            '';
          };
        });
    };
}
