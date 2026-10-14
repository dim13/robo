// Decompile every function to one C file.
import ghidra.app.decompiler.*;
import ghidra.app.script.GhidraScript;
import ghidra.program.model.listing.*;
import java.io.PrintWriter;

public class Dump extends GhidraScript {
	@Override
	public void run() throws Exception {
		String out = getScriptArgs().length > 0 ? getScriptArgs()[0] : "/tmp/out.c";
		DecompInterface d = new DecompInterface();
		d.openProgram(currentProgram);
		int n = 0, bad = 0;
		try (PrintWriter w = new PrintWriter(out)) {
			for (Function f : currentProgram.getFunctionManager().getFunctions(true)) {
				DecompileResults r = d.decompileFunction(f, 120, monitor);
				n++;
				if (r.decompileCompleted()) {
					w.println("// " + f.getEntryPoint());
					w.println(r.getDecompiledFunction().getC());
				} else {
					bad++;
					w.println("// " + f.getEntryPoint() + " " + f.getName() + ": " + r.getErrorMessage());
				}
			}
		}
		println("functions " + n + " failed " + bad);
	}
}
