// Apply symbols file: "addr name [C prototype]" for code, "addr name [type]" for data; '#' comments.
import ghidra.app.cmd.function.ApplyFunctionSignatureCmd;
import ghidra.app.script.GhidraScript;
import ghidra.app.util.cparser.C.CParserUtils;
import ghidra.program.model.address.Address;
import ghidra.program.model.data.*;
import ghidra.program.model.listing.Function;
import ghidra.program.model.symbol.SourceType;
import ghidra.util.data.DataTypeParser;
import java.nio.file.*;

public class Names extends GhidraScript {
	@Override
	public void run() throws Exception {
		DataTypeParser dp = new DataTypeParser(currentProgram.getDataTypeManager(), null, null, DataTypeParser.AllowedDataTypes.ALL);
		for (String line : Files.readAllLines(Path.of(getScriptArgs()[0]))) {
			line = line.replaceAll("#.*", "").trim();
			if (line.isEmpty()) continue;
			String[] f = line.split("\\s+", 3);
			Address a = toAddr(Long.parseLong(f[0], 16));
			try {
				Function fn = getFunctionAt(a);
				if (fn == null && a.getOffset() < 0x40000 && f.length > 2) {
					disassemble(a);
					fn = createFunction(a, f[1]);
				}
				if (fn != null) {
					fn.setName(f[1], SourceType.USER_DEFINED);
					if (f.length > 2) {
						FunctionDefinitionDataType sig = new ghidra.app.util.parser.FunctionSignatureParser(currentProgram.getDataTypeManager(), null).parse(null, f[2].replaceAll(";$", ""));
						if (sig != null) new ApplyFunctionSignatureCmd(a, sig, SourceType.USER_DEFINED).applyTo(currentProgram);
						else println("bad sig " + line);
					}
					continue;
				}
				createLabel(a, f[1], true, SourceType.USER_DEFINED);
				if (f.length > 2) {
					DataType t = dp.parse(f[2]);
					clearListing(a, a.add(t.getLength() - 1));
					createData(a, t);
				}
			} catch (Exception e) {
				println("fail " + line + ": " + e);
			}
		}
	}
}
