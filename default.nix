with import <nixpkgs> { };

buildGoModule {
  name = "gh";
  src = lib.cleanSource ./.;
  vendorSha256 = "0w9km75mr41k9mqi8cc33vm0lwrppqbdi7vwma68q8gvyyidz8x7";
}
