// Print disassembly: args start end
import ghidra.app.script.GhidraScript;
import ghidra.program.model.listing.*;
public class Dis extends GhidraScript {
	@Override
	public void run() throws Exception {
		String[] a = getScriptArgs();
		var it = currentProgram.getListing().getInstructions(currentProgram.getAddressFactory().getAddressSet(toAddr(Long.decode(a[0])), toAddr(Long.decode(a[1]))), true);
		StringBuilder sb = new StringBuilder();
		while (it.hasNext()) { Instruction i = it.next(); sb.append(i.getAddress()).append("  ").append(i).append("\n"); }
		println("\n" + sb);
	}
}
