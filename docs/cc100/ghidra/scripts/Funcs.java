import ghidra.app.script.GhidraScript;
import ghidra.program.model.listing.*;
import java.io.PrintWriter;
public class Funcs extends GhidraScript {
	@Override
	public void run() throws Exception {
		try (PrintWriter w = new PrintWriter(getScriptArgs()[0])) {
			for (Function f : currentProgram.getFunctionManager().getFunctions(true)) {
				int callers = f.getCallingFunctions(monitor).size();
				int callees = f.getCalledFunctions(monitor).size();
				w.printf("%s %6d %3d %3d %s%n", f.getEntryPoint(), f.getBody().getNumAddresses(), callers, callees, f.getName());
			}
		}
	}
}
