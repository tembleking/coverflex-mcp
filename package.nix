{
  lib,
  buildGoModule,
}:
buildGoModule rec {
  pname = "coverflex-mcp";
  version = "0.0.1";

  src = ./.;

  vendorHash = "sha256-v60oPY65xqAQgEPg4AE8uNiliQCFxdOBxj7Wyl1Q4r0=";

  meta = with lib; {
    description = "MCP server for Coverflex";
    homepage = "https://github.com/tembleking/coverflex-mcp";
    license = licenses.asl20;
    maintainers = with maintainers; [ tembleking ];
  };
}
