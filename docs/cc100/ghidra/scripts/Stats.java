import ghidra.app.script.GhidraScript;
import ghidra.program.model.listing.*;
import ghidra.program.model.address.*;
import java.util.*;

public class Stats extends GhidraScript {
	@Override
	public void run() throws Exception {
		Listing l = currentProgram.getListing();
		long ins = 0;
		for (Instruction i : l.getInstructions(true)) ins += i.getLength();
		println("insn bytes " + ins);
		Map<String,Integer> m = new TreeMap<>();
		var it = currentProgram.getBookmarkManager().getBookmarksIterator("Error");
		while (it.hasNext()) { var b = it.next();
			String k = b.getComment().replaceAll("at .*","");
			byte[] by = new byte[8];
			try { currentProgram.getMemory().getBytes(b.getAddress(), by); } catch (Exception e) {}
			k = String.format("%02x%02x%02x%02x", by[0],by[1],by[2],by[3]);
			m.merge(k.substring(0,4), 1, Integer::sum); if (k.startsWith("0100")) { byte[] q=new byte[8]; try{currentProgram.getMemory().getBytes(b.getAddress(), q);}catch(Exception e){} StringBuilder sb=new StringBuilder(); for(byte x:q) sb.append(String.format("%02x",x)); println(b.getAddress()+" "+sb); }
		}
		println("errors " + m);
	}
}
