When planning or building code implementation, follows this:
- Use Go idiomatic.
- Allocate as less memory as possible and try to not put pressure on the GC.
- Use if guard clause at the beginning of function or loop.
