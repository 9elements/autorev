# IR

IR is a representation of high level code.

Currently 3 types exist: Temporaries, Primitives and Complex

## Temporaries
Temporaries (T) are variables used to store data read, to apply logical operations
and to write data.

## Primitives
The simplest IR instructins are primitives.
A primitive (P) is a read, a write operation or a logical operation on a temporary.
Example:
  read32((void *)0x500) shown as P:read32
  write8((void *)0xdeadbeef, 0) shown as P:write8

## Complex
Everything not a primitive is called complex (C).
Complex have a higher rank and can overwrite IR instructions with a lower rank.
Complex contain a list of primitives.

**Read-Modify-Write (RMW):**
 * Scans for P:readX, P:writeX
 * Replaces P:readX, p:writeX
 * Needs to gather the & and | mask by running additional test

**Branch on bit set (BBS):**
 * Scans for branch after P:readX
 * Replaceds P:readX
 * Needs to gather the bit set by running additional test
 * Adds a condition to the AST

**Branch on bit clear (BBC):**
 * Scans for branch after P:readX
 * Replaceds P:readX
 * Needs to gather the bit cleared by running additional test
 * Adds a condition to the AST

**Branch on masked integer (BMI):**
 * Scans for branch after P:readX
 * Replaceds P:readX
 * Needs to gather the mask and integer by running additional test
 * Adds a condition to the AST

**Simple loop detection counter (SLDC):**
 * Scans for duplicated primitives
 * Replaces primitives
 * Includes a counter

**Simple loop detection bit set (SLDBS):**
 * Scans for duplicated primitives
 * The primitives contain a P:readX
 * Replaces primitives
 * Checks for bit set

**Simple loop detection bit clear (SLDBC):**
 * Scans for duplicated primitives
 * The primitives contain a P:readX
 * Replaces primitives
 * Checks for bit clear

## Conversion from tracelog

The tracelog contains tracelog entries. A tracelog entry can only be converted
to a primitive.

The IR AST is generated from all existing tracelogs.

Then the IR AST plugins are run and scan the AST for possible "optimizations".
The lower rank plugins are run first, then the higher ones.
