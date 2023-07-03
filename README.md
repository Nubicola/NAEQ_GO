# NAEQ_GO

This command-line utility will read inputs and output eq values. The purpose is to quickly build NAEQ_##.md files for use in the other project, naeq_obsidian. 

In particular,
-f=file to read a file
-d=dir  to read all files in a directory
(list on the command line) to process the words on the command line.

-o=dir to write NAEQ_##.md files. It will merge existing content it finds there.

-p=mode, where mode is "word" (default), "line" or "sent". This affects the target of the calculation: EQ on a word basis, a line (\n) basis, or a sentence basis
Note that "sentence" basis is computationally expensive. Processing large texts in that way is not recommended!

There's not a ton of error checking or resiliance. 


