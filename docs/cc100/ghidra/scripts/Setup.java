// Pre-analysis setup: RAM/IO block and vector table entries.
import ghidra.app.script.GhidraScript;
import ghidra.program.model.address.Address;
import ghidra.program.model.data.PointerDataType;
import ghidra.program.model.mem.MemoryBlock;
import ghidra.program.model.symbol.SourceType;

public class Setup extends GhidraScript {
	@Override
	public void run() throws Exception {
		MemoryBlock b = currentProgram.getMemory().createUninitializedBlock("RAM", toAddr(0xff0000), 0x10000, false);
		b.setWrite(true);
		b.setVolatile(true);
		for (int i = 0; i < 128; i++) {
			Address va = toAddr(i * 4);
			long t = getInt(va) & 0xffffffffL;
			if (t == 0xffffffffL || t >= 0x40000) continue;
			createData(va, PointerDataType.dataType);
			Address ta = toAddr(t);
			String name = i == 0 ? "reset" : "vec" + i;
			createLabel(ta, name, true, SourceType.USER_DEFINED);
			addEntryPoint(ta);
			disassemble(ta);
			createFunction(ta, name);
		}
	}
}
