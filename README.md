# go-hash-walk

This project is setup to test how its possible to speed up the generation of
thousands of hashes for a directory (tree) of files.

The inspiration for looking into this is the [Witness](https://witness.dev)
project that has generates hashes at the beginning and end of a command run
to capture the state of a whole working tree.

This process can take a long time if its only done one file at a time. This
implementation of that process indexes the working tree and start a worker
process in the backgrond to generate all the hashes.
