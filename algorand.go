package teal

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// The code below is licensed under the following conditions: (https://github.com/algorand/go-algorand-sdk/blob/develop/LICENSE)

/*
MIT License

Copyright (c) 2019 Algorand, llc

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

const protoByte = 0x8a

// directRefEnabledVersion is the version of TEAL where opcodes
// that reference accounts, asas, and apps may do so directly, not requiring
// using an index into arrays.
const directRefEnabledVersion = 4

const fidoVersion = 7       // base64, json, secp256r1
const randomnessVersion = 7 // vrf_verify, block
const fpVersion = 8         // changes for frame pointers and simpler function discipline

// EXPERIMENTAL. These should be revisited whenever a new LogicSigVersion is
// moved from vFuture to a new consensus version. If they remain unready, bump
// their version, and fixup TestAssemble() in assembler_test.go.
const pairingVersion = 9 // bn256 opcodes. will add bls12-381, and unify the available opcodes.

// Unlimited Global Storage opcodes
const boxVersion = 8 // box_*

const anyImmediates = -1

// short description of every op
var opDocByName = map[string]string{
	"err":                 "Fail immediately.",
	"sha256":              "SHA256 hash of value A, yields [32]byte",
	"keccak256":           "Keccak256 hash of value A, yields [32]byte",
	"sha512_256":          "SHA512_256 hash of value A, yields [32]byte",
	"sha3_256":            "SHA3_256 hash of value A, yields [32]byte",
	"ed25519verify":       "for (data A, signature B, pubkey C) verify the signature of (\"ProgData\" || program_hash || data) against the pubkey => {0 or 1}",
	"ed25519verify_bare":  "for (data A, signature B, pubkey C) verify the signature of the data against the pubkey => {0 or 1}",
	"ecdsa_verify":        "for (data A, signature B, C and pubkey D, E) verify the signature of the data against the pubkey => {0 or 1}",
	"ecdsa_pk_decompress": "decompress pubkey A into components X, Y",
	"ecdsa_pk_recover":    "for (data A, recovery id B, signature C, D) recover a public key",
	"bn256_add":           "for (curve points A and B) return the curve point A + B",
	"bn256_scalar_mul":    "for (curve point A, scalar K) return the curve point KA",
	"bn256_pairing":       "for (points in G1 group G1s, points in G2 group G2s), return whether they are paired => {0 or 1}",

	"+":       "A plus B. Fail on overflow.",
	"-":       "A minus B. Fail if B > A.",
	"/":       "A divided by B (truncated division). Fail if B == 0.",
	"*":       "A times B. Fail on overflow.",
	"<":       "A less than B => {0 or 1}",
	">":       "A greater than B => {0 or 1}",
	"<=":      "A less than or equal to B => {0 or 1}",
	">=":      "A greater than or equal to B => {0 or 1}",
	"&&":      "A is not zero and B is not zero => {0 or 1}",
	"||":      "A is not zero or B is not zero => {0 or 1}",
	"==":      "A is equal to B => {0 or 1}",
	"!=":      "A is not equal to B => {0 or 1}",
	"!":       "A == 0 yields 1; else 0",
	"len":     "yields length of byte value A",
	"itob":    "converts uint64 A to big-endian byte array, always of length 8",
	"btoi":    "converts big-endian byte array A to uint64. Fails if len(A) > 8. Padded by leading 0s if len(A) < 8.",
	"%":       "A modulo B. Fail if B == 0.",
	"|":       "A bitwise-or B",
	"&":       "A bitwise-and B",
	"^":       "A bitwise-xor B",
	"~":       "bitwise invert value A",
	"shl":     "A times 2^B, modulo 2^64",
	"shr":     "A divided by 2^B",
	"sqrt":    "The largest integer I such that I^2 <= A",
	"bitlen":  "The highest set bit in A. If A is a byte-array, it is interpreted as a big-endian unsigned integer. bitlen of 0 is 0, bitlen of 8 is 4",
	"exp":     "A raised to the Bth power. Fail if A == B == 0 and on overflow",
	"expw":    "A raised to the Bth power as a 128-bit result in two uint64s. X is the high 64 bits, Y is the low. Fail if A == B == 0 or if the results exceeds 2^128-1",
	"mulw":    "A times B as a 128-bit result in two uint64s. X is the high 64 bits, Y is the low",
	"addw":    "A plus B as a 128-bit result. X is the carry-bit, Y is the low-order 64 bits.",
	"divw":    "A,B / C. Fail if C == 0 or if result overflows.",
	"divmodw": "W,X = (A,B / C,D); Y,Z = (A,B modulo C,D)",

	"intcblock":  "prepare block of uint64 constants for use by intc",
	"intc":       "Ith constant from intcblock",
	"intc_0":     "constant 0 from intcblock",
	"intc_1":     "constant 1 from intcblock",
	"intc_2":     "constant 2 from intcblock",
	"intc_3":     "constant 3 from intcblock",
	"pushint":    "immediate UINT",
	"pushints":   "push sequence of immediate uints to stack in the order they appear (first uint being deepest)",
	"bytecblock": "prepare block of byte-array constants for use by bytec",
	"bytec":      "Ith constant from bytecblock",
	"bytec_0":    "constant 0 from bytecblock",
	"bytec_1":    "constant 1 from bytecblock",
	"bytec_2":    "constant 2 from bytecblock",
	"bytec_3":    "constant 3 from bytecblock",
	"pushbytes":  "immediate BYTES",
	"pushbytess": "push sequences of immediate byte arrays to stack (first byte array being deepest)",

	"bzero":   "zero filled byte-array of length A",
	"arg":     "Nth LogicSig argument",
	"arg_0":   "LogicSig argument 0",
	"arg_1":   "LogicSig argument 1",
	"arg_2":   "LogicSig argument 2",
	"arg_3":   "LogicSig argument 3",
	"args":    "Ath LogicSig argument",
	"txn":     "field F of current transaction",
	"gtxn":    "field F of the Tth transaction in the current group",
	"gtxns":   "field F of the Ath transaction in the current group",
	"txna":    "Ith value of the array field F of the current transaction",
	"gtxna":   "Ith value of the array field F from the Tth transaction in the current group",
	"gtxnsa":  "Ith value of the array field F from the Ath transaction in the current group",
	"txnas":   "Ath value of the array field F of the current transaction",
	"gtxnas":  "Ath value of the array field F from the Tth transaction in the current group",
	"gtxnsas": "Bth value of the array field F from the Ath transaction in the current group",
	"itxn":    "field F of the last inner transaction",
	"itxna":   "Ith value of the array field F of the last inner transaction",
	"itxnas":  "Ath value of the array field F of the last inner transaction",
	"gitxn":   "field F of the Tth transaction in the last inner group submitted",
	"gitxna":  "Ith value of the array field F from the Tth transaction in the last inner group submitted",
	"gitxnas": "Ath value of the array field F from the Tth transaction in the last inner group submitted",

	"global":  "global field F",
	"load":    "Ith scratch space value. All scratch spaces are 0 at program start.",
	"store":   "store A to the Ith scratch space",
	"loads":   "Ath scratch space value.  All scratch spaces are 0 at program start.",
	"stores":  "store B to the Ath scratch space",
	"gload":   "Ith scratch space value of the Tth transaction in the current group",
	"gloads":  "Ith scratch space value of the Ath transaction in the current group",
	"gloadss": "Bth scratch space value of the Ath transaction in the current group",
	"gaid":    "ID of the asset or application created in the Tth transaction of the current group",
	"gaids":   "ID of the asset or application created in the Ath transaction of the current group",

	"json_ref": "key B's value, of type R, from a [valid](jsonspec.md) utf-8 encoded json object A",

	"bnz":     "branch to TARGET if value A is not zero",
	"bz":      "branch to TARGET if value A is zero",
	"b":       "branch unconditionally to TARGET",
	"return":  "use A as success value; end",
	"pop":     "discard A",
	"dup":     "duplicate A",
	"dup2":    "duplicate A and B",
	"dupn":    "duplicate A, N times",
	"dig":     "Nth value from the top of the stack. dig 0 is equivalent to dup",
	"bury":    "replace the Nth value from the top of the stack with A. bury 0 fails.",
	"cover":   "remove top of stack, and place it deeper in the stack such that N elements are above it. Fails if stack depth <= N.",
	"uncover": "remove the value at depth N in the stack and shift above items down so the Nth deep value is on top of the stack. Fails if stack depth <= N.",
	"swap":    "swaps A and B on stack",
	"select":  "selects one of two values based on top-of-stack: B if C != 0, else A",

	"concat":            "join A and B",
	"substring":         "A range of bytes from A starting at S up to but not including E. If E < S, or either is larger than the array length, the program fails",
	"substring3":        "A range of bytes from A starting at B up to but not including C. If C < B, or either is larger than the array length, the program fails",
	"getbit":            "Bth bit of (byte-array or integer) A. If B is greater than or equal to the bit length of the value (8*byte length), the program fails",
	"setbit":            "Copy of (byte-array or integer) A, with the Bth bit set to (0 or 1) C. If B is greater than or equal to the bit length of the value (8*byte length), the program fails",
	"getbyte":           "Bth byte of A, as an integer. If B is greater than or equal to the array length, the program fails",
	"setbyte":           "Copy of A with the Bth byte set to small integer (between 0..255) C. If B is greater than or equal to the array length, the program fails",
	"extract":           "A range of bytes from A starting at S up to but not including S+L. If L is 0, then extract to the end of the string. If S or S+L is larger than the array length, the program fails",
	"extract3":          "A range of bytes from A starting at B up to but not including B+C. If B+C is larger than the array length, the program fails",
	"extract_uint16":    "A uint16 formed from a range of big-endian bytes from A starting at B up to but not including B+2. If B+2 is larger than the array length, the program fails",
	"extract_uint32":    "A uint32 formed from a range of big-endian bytes from A starting at B up to but not including B+4. If B+4 is larger than the array length, the program fails",
	"extract_uint64":    "A uint64 formed from a range of big-endian bytes from A starting at B up to but not including B+8. If B+8 is larger than the array length, the program fails",
	"replace2":          "Copy of A with the bytes starting at S replaced by the bytes of B. Fails if S+len(B) exceeds len(A)",
	"replace3":          "Copy of A with the bytes starting at B replaced by the bytes of C. Fails if B+len(C) exceeds len(A)",
	"base64_decode":     "decode A which was base64-encoded using _encoding_ E. Fail if A is not base64 encoded with encoding E",
	"balance":           "balance for account A, in microalgos. The balance is observed after the effects of previous transactions in the group, and after the fee for the current transaction is deducted. Changes caused by inner transactions are observable immediately following `itxn_submit`",
	"min_balance":       "minimum required balance for account A, in microalgos. Required balance is affected by ASA, App, and Box usage. When creating or opting into an app, the minimum balance grows before the app code runs, therefore the increase is visible there. When deleting or closing out, the minimum balance decreases after the app executes. Changes caused by inner transactions or box usage are observable immediately following the opcode effecting the change.",
	"app_opted_in":      "1 if account A is opted in to application B, else 0",
	"app_local_get":     "local state of the key B in the current application in account A",
	"app_local_get_ex":  "X is the local state of application B, key C in account A. Y is 1 if key existed, else 0",
	"app_global_get":    "global state of the key A in the current application",
	"app_global_get_ex": "X is the global state of application A, key B. Y is 1 if key existed, else 0",
	"app_local_put":     "write C to key B in account A's local state of the current application",
	"app_global_put":    "write B to key A in the global state of the current application",
	"app_local_del":     "delete key B from account A's local state of the current application",
	"app_global_del":    "delete key A from the global state of the current application",
	"asset_holding_get": "X is field F from account A's holding of asset B. Y is 1 if A is opted into B, else 0",
	"asset_params_get":  "X is field F from asset A. Y is 1 if A exists, else 0",
	"app_params_get":    "X is field F from app A. Y is 1 if A exists, else 0",
	"acct_params_get":   "X is field F from account A. Y is 1 if A owns positive algos, else 0",
	"assert":            "immediately fail unless A is a non-zero number",
	"callsub":           "branch unconditionally to TARGET, saving the next instruction on the call stack",
	"proto":             "Prepare top call frame for a retsub that will assume A args and R return values.",
	"retsub":            "pop the top instruction from the call stack and branch to it",

	"b+":  "A plus B. A and B are interpreted as big-endian unsigned integers",
	"b-":  "A minus B. A and B are interpreted as big-endian unsigned integers. Fail on underflow.",
	"b/":  "A divided by B (truncated division). A and B are interpreted as big-endian unsigned integers. Fail if B is zero.",
	"b*":  "A times B. A and B are interpreted as big-endian unsigned integers.",
	"b<":  "1 if A is less than B, else 0. A and B are interpreted as big-endian unsigned integers",
	"b>":  "1 if A is greater than B, else 0. A and B are interpreted as big-endian unsigned integers",
	"b<=": "1 if A is less than or equal to B, else 0. A and B are interpreted as big-endian unsigned integers",
	"b>=": "1 if A is greater than or equal to B, else 0. A and B are interpreted as big-endian unsigned integers",
	"b==": "1 if A is equal to B, else 0. A and B are interpreted as big-endian unsigned integers",
	"b!=": "0 if A is equal to B, else 1. A and B are interpreted as big-endian unsigned integers",
	"b%":  "A modulo B. A and B are interpreted as big-endian unsigned integers. Fail if B is zero.",
	"b|":  "A bitwise-or B. A and B are zero-left extended to the greater of their lengths",
	"b&":  "A bitwise-and B. A and B are zero-left extended to the greater of their lengths",
	"b^":  "A bitwise-xor B. A and B are zero-left extended to the greater of their lengths",
	"b~":  "A with all bits inverted",

	"bsqrt": "The largest integer I such that I^2 <= A. A and I are interpreted as big-endian unsigned integers",

	"log":         "write A to log state of the current application",
	"itxn_begin":  "begin preparation of a new inner transaction in a new transaction group",
	"itxn_next":   "begin preparation of a new inner transaction in the same transaction group",
	"itxn_field":  "set field F of the current inner transaction to A",
	"itxn_submit": "execute the current inner transaction group. Fail if executing this group would exceed the inner transaction limit, or if any transaction in the group fails.",

	"vrf_verify": "Verify the proof B of message A against pubkey C. Returns vrf output and verification flag.",
	"block":      "field F of block A. Fail unless A falls between txn.LastValid-1002 and txn.FirstValid (exclusive)",

	"switch": "branch to the Ath label. Continue at following instruction if index A exceeds the number of labels.",
	"match":  "given match cases from A[1] to A[N], branch to the Ith label where A[I] = B. Continue to the following instruction if no matches are found.",

	"frame_dig":  "Nth (signed) value from the frame pointer.",
	"frame_bury": "replace the Nth (signed) value from the frame pointer in the stack with A",
	"popn":       "remove N values from the top of the stack",

	"box_create":  "create a box named A, of length B. Fail if A is empty or B exceeds 32,768. Returns 0 if A already existed, else 1",
	"box_extract": "read C bytes from box A, starting at offset B. Fail if A does not exist, or the byte range is outside A's size.",
	"box_replace": "write byte-array C into box A, starting at offset B. Fail if A does not exist, or the byte range is outside A's size.",
	"box_del":     "delete box named A if it exists. Return 1 if A existed, 0 otherwise",
	"box_len":     "X is the length of box A if A exists, else 0. Y is 1 if A exists, else 0.",
	"box_get":     "X is the contents of box A if A exists, else ''. Y is 1 if A exists, else 0.",
	"box_put":     "replaces the contents of box A with byte-array B. Fails if A exists and len(B) != len(box A). Creates A if it does not exist",
}

var opDocExtras = map[string]string{
	"vrf_verify":          "`VrfAlgorand` is the VRF used in Algorand. It is ECVRF-ED25519-SHA512-Elligator2, specified in the IETF internet draft [draft-irtf-cfrg-vrf-03](https://datatracker.ietf.org/doc/draft-irtf-cfrg-vrf/03/).",
	"ed25519verify":       "The 32 byte public key is the last element on the stack, preceded by the 64 byte signature at the second-to-last element on the stack, preceded by the data which was signed at the third-to-last element on the stack.",
	"ecdsa_verify":        "The 32 byte Y-component of a public key is the last element on the stack, preceded by X-component of a pubkey, preceded by S and R components of a signature, preceded by the data that is fifth element on the stack. All values are big-endian encoded. The signed data must be 32 bytes long, and signatures in lower-S form are only accepted.",
	"ecdsa_pk_decompress": "The 33 byte public key in a compressed form to be decompressed into X and Y (top) components. All values are big-endian encoded.",
	"ecdsa_pk_recover":    "S (top) and R elements of a signature, recovery id and data (bottom) are expected on the stack and used to deriver a public key. All values are big-endian encoded. The signed data must be 32 bytes long.",
	"bn256_add":           "A, B are curve points in G1 group. Each point consists of (X, Y) where X and Y are 256 bit integers, big-endian encoded. The encoded point is 64 bytes from concatenation of 32 byte X and 32 byte Y.",
	"bn256_scalar_mul":    "A is a curve point in G1 Group and encoded as described in `bn256_add`. Scalar K is a big-endian encoded big integer that has no padding zeros.",
	"bn256_pairing":       "G1s are encoded by the concatenation of encoded G1 points, as described in `bn256_add`. G2s are encoded by the concatenation of encoded G2 points. Each G2 is in form (XA0+i*XA1, YA0+i*YA1) and encoded by big-endian field element XA0, XA1, YA0 and YA1 in sequence.",
	"bnz":                 "The `bnz` instruction opcode 0x40 is followed by two immediate data bytes which are a high byte first and low byte second which together form a 16 bit offset which the instruction may branch to. For a bnz instruction at `pc`, if the last element of the stack is not zero then branch to instruction at `pc + 3 + N`, else proceed to next instruction at `pc + 3`. Branch targets must be aligned instructions. (e.g. Branching to the second byte of a 2 byte op will be rejected.) Starting at v4, the offset is treated as a signed 16 bit integer allowing for backward branches and looping. In prior version (v1 to v3), branch offsets are limited to forward branches only, 0-0x7fff.\n\nAt v2 it became allowed to branch to the end of the program exactly after the last instruction: bnz to byte N (with 0-indexing) was illegal for a TEAL program with N bytes before v2, and is legal after it. This change eliminates the need for a last instruction of no-op as a branch target at the end. (Branching beyond the end--in other words, to a byte larger than N--is still illegal and will cause the program to fail.)",
	"bz":                  "See `bnz` for details on how branches work. `bz` inverts the behavior of `bnz`.",
	"b":                   "See `bnz` for details on how branches work. `b` always jumps to the offset.",
	"callsub":             "The call stack is separate from the data stack. Only `callsub`, `retsub`, and `proto` manipulate it.",
	"proto":               "Fails unless the last instruction executed was a `callsub`.",
	"retsub":              "If the current frame was prepared by `proto A R`, `retsub` will remove the 'A' arguments from the stack, move the `R` return values down, and pop any stack locations above the relocated return values.",
	"intcblock":           "`intcblock` loads following program bytes into an array of integer constants in the evaluator. These integer constants can be referred to by `intc` and `intc_*` which will push the value onto the stack. Subsequent calls to `intcblock` reset and replace the integer constants available to the script.",
	"bytecblock":          "`bytecblock` loads the following program bytes into an array of byte-array constants in the evaluator. These constants can be referred to by `bytec` and `bytec_*` which will push the value onto the stack. Subsequent calls to `bytecblock` reset and replace the bytes constants available to the script.",
	"*":                   "Overflow is an error condition which halts execution and fails the transaction. Full precision is available from `mulw`.",
	"+":                   "Overflow is an error condition which halts execution and fails the transaction. Full precision is available from `addw`.",
	"/":                   "`divmodw` is available to divide the two-element values produced by `mulw` and `addw`.",
	"bitlen":              "bitlen interprets arrays as big-endian integers, unlike setbit/getbit",
	"divw":                "The notation A,B indicates that A and B are interpreted as a uint128 value, with A as the high uint64 and B the low.",
	"divmodw":             "The notation J,K indicates that two uint64 values J and K are interpreted as a uint128 value, with J as the high uint64 and K the low.",
	"gtxn":                "for notes on transaction fields available, see `txn`. If this transaction is _i_ in the group, `gtxn i field` is equivalent to `txn field`.",
	"gtxns":               "for notes on transaction fields available, see `txn`. If top of stack is _i_, `gtxns field` is equivalent to `gtxn _i_ field`. gtxns exists so that _i_ can be calculated, often based on the index of the current transaction.",
	"gload":               "`gload` fails unless the requested transaction is an ApplicationCall and T < GroupIndex.",
	"gloads":              "`gloads` fails unless the requested transaction is an ApplicationCall and A < GroupIndex.",
	"gaid":                "`gaid` fails unless the requested transaction created an asset or application and T < GroupIndex.",
	"gaids":               "`gaids` fails unless the requested transaction created an asset or application and A < GroupIndex.",
	"btoi":                "`btoi` fails if the input is longer than 8 bytes.",
	"concat":              "`concat` fails if the result would be greater than 4096 bytes.",
	"pushbytes":           "pushbytes args are not added to the bytecblock during assembly processes",
	"pushbytess":          "pushbytess args are not added to the bytecblock during assembly processes",
	"pushint":             "pushint args are not added to the intcblock during assembly processes",
	"pushints":            "pushints args are not added to the intcblock during assembly processes",
	"getbit":              "see explanation of bit ordering in setbit",
	"setbit":              "When A is a uint64, index 0 is the least significant bit. Setting bit 3 to 1 on the integer 0 yields 8, or 2^3. When A is a byte array, index 0 is the leftmost bit of the leftmost byte. Setting bits 0 through 11 to 1 in a 4-byte-array of 0s yields the byte array 0xfff00000. Setting bit 3 to 1 on the 1-byte-array 0x00 yields the byte array 0x10.",
	"balance":             "params: Txn.Accounts offset (or, since v4, an _available_ account address), _available_ application id (or, since v4, a Txn.ForeignApps offset). Return: value.",
	"min_balance":         "params: Txn.Accounts offset (or, since v4, an _available_ account address), _available_ application id (or, since v4, a Txn.ForeignApps offset). Return: value.",
	"app_opted_in":        "params: Txn.Accounts offset (or, since v4, an _available_ account address), _available_ application id (or, since v4, a Txn.ForeignApps offset). Return: 1 if opted in and 0 otherwise.",
	"app_local_get":       "params: Txn.Accounts offset (or, since v4, an _available_ account address), state key. Return: value. The value is zero (of type uint64) if the key does not exist.",
	"app_local_get_ex":    "params: Txn.Accounts offset (or, since v4, an _available_ account address), _available_ application id (or, since v4, a Txn.ForeignApps offset), state key. Return: did_exist flag (top of the stack, 1 if the application and key existed and 0 otherwise), value. The value is zero (of type uint64) if the key does not exist.",
	"app_global_get_ex":   "params: Txn.ForeignApps offset (or, since v4, an _available_ application id), state key. Return: did_exist flag (top of the stack, 1 if the application and key existed and 0 otherwise), value. The value is zero (of type uint64) if the key does not exist.",
	"app_global_get":      "params: state key. Return: value. The value is zero (of type uint64) if the key does not exist.",
	"app_local_put":       "params: Txn.Accounts offset (or, since v4, an _available_ account address), state key, value.",
	"app_local_del":       "params: Txn.Accounts offset (or, since v4, an _available_ account address), state key.\n\nDeleting a key which is already absent has no effect on the application local state. (In particular, it does _not_ cause the program to fail.)",
	"app_global_del":      "params: state key.\n\nDeleting a key which is already absent has no effect on the application global state. (In particular, it does _not_ cause the program to fail.)",
	"asset_holding_get":   "params: Txn.Accounts offset (or, since v4, an _available_ address), asset id (or, since v4, a Txn.ForeignAssets offset). Return: did_exist flag (1 if the asset existed and 0 otherwise), value.",
	"asset_params_get":    "params: Txn.ForeignAssets offset (or, since v4, an _available_ asset id. Return: did_exist flag (1 if the asset existed and 0 otherwise), value.",
	"app_params_get":      "params: Txn.ForeignApps offset or an _available_ app id. Return: did_exist flag (1 if the application existed and 0 otherwise), value.",
	"log":                 "`log` fails if called more than MaxLogCalls times in a program, or if the sum of logged bytes exceeds 1024 bytes.",
	"itxn_begin":          "`itxn_begin` initializes Sender to the application address; Fee to the minimum allowable, taking into account MinTxnFee and credit from overpaying in earlier transactions; FirstValid/LastValid to the values in the invoking transaction, and all other fields to zero or empty values.",
	"itxn_next":           "`itxn_next` initializes the transaction exactly as `itxn_begin` does",
	"itxn_field":          "`itxn_field` fails if A is of the wrong type for F, including a byte array of the wrong size for use as an address when F is an address field. `itxn_field` also fails if A is an account, asset, or app that is not _available_, or an attempt is made extend an array field beyond the limit imposed by consensus parameters. (Addresses set into asset params of acfg transactions need not be _available_.)",
	"itxn_submit":         "`itxn_submit` resets the current transaction so that it can not be resubmitted. A new `itxn_begin` is required to prepare another inner transaction.",

	"base64_decode": "*Warning*: Usage should be restricted to very rare use cases. In almost all cases, smart contracts should directly handle non-encoded byte-strings.	This opcode should only be used in cases where base64 is the only available option, e.g. interoperability with a third-party that only signs base64 strings.\n\n Decodes A using the base64 encoding E. Specify the encoding with an immediate arg either as URL and Filename Safe (`URLEncoding`) or Standard (`StdEncoding`). See [RFC 4648 sections 4 and 5](https://rfc-editor.org/rfc/rfc4648.html#section-4). It is assumed that the encoding ends with the exact number of `=` padding characters as required by the RFC. When padding occurs, any unused pad bits in the encoding must be set to zero or the decoding will fail. The special cases of `\\n` and `\\r` are allowed but completely ignored. An error will result when attempting to decode a string with a character that is not in the encoding alphabet or not one of `=`, `\\r`, or `\\n`.",
	"json_ref":      "*Warning*: Usage should be restricted to very rare use cases, as JSON decoding is expensive and quite limited. In addition, JSON objects are large and not optimized for size.\n\nAlmost all smart contracts should use simpler and smaller methods (such as the [ABI](https://arc.algorand.foundation/ARCs/arc-0004). This opcode should only be used in cases where JSON is only available option, e.g. when a third-party only signs JSON.",

	"match": "`match` consumes N+1 values from the stack. Let the top stack value be B. The following N values represent an ordered list of match cases/constants (A), where the first value (A[0]) is the deepest in the stack. The immediate arguments are an ordered list of N labels (T). `match` will branch to target T[I], where A[I] = B. If there are no matches then execution continues on to the next instruction.",

	"box_create": "Newly created boxes are filled with 0 bytes. `box_create` will fail if the referenced box already exists with a different size. Otherwise, existing boxes are unchanged by `box_create`.",
	"box_get":    "For boxes that exceed 4,096 bytes, consider `box_create`, `box_extract`, and `box_replace`",
	"box_put":    "For boxes that exceed 4,096 bytes, consider `box_create`, `box_extract`, and `box_replace`",
}

var pseudoOps = map[string]map[int]OpSpec{
	"int":  {anyImmediates: OpSpec{Name: "int", Proto: proto(":i"), OpDetails: assembler(asmInt)}},
	"byte": {anyImmediates: OpSpec{Name: "byte", Proto: proto(":b"), OpDetails: assembler(asmByte)}},
	// parse basics.Address, actually just another []byte constant
	"addr": {anyImmediates: OpSpec{Name: "addr", Proto: proto(":b"), OpDetails: assembler(asmAddr)}},
	// take a signature, hash it, and take first 4 bytes, actually just another []byte constant
	"method":  {anyImmediates: OpSpec{Name: "method", Proto: proto(":b"), OpDetails: assembler(asmMethod)}},
	"txn":     {1: OpSpec{Name: "txn"}, 2: OpSpec{Name: "txna"}},
	"gtxn":    {2: OpSpec{Name: "gtxn"}, 3: OpSpec{Name: "gtxna"}},
	"gtxns":   {1: OpSpec{Name: "gtxns"}, 2: OpSpec{Name: "gtxnsa"}},
	"extract": {0: OpSpec{Name: "extract3"}, 2: OpSpec{Name: "extract"}},
	"replace": {0: OpSpec{Name: "replace3"}, 1: OpSpec{Name: "replace2"}},
}

var OpSpecs = []OpSpec{
	{0x00, "err", opErr, proto(":x"), 1, detDefault()},
	{0x01, "sha256", opSHA256, proto("b:b"), 1, costly(7)},
	{0x02, "keccak256", opKeccak256, proto("b:b"), 1, costly(26)},
	{0x03, "sha512_256", opSHA512_256, proto("b:b"), 1, costly(9)},

	// Cost of these opcodes increases in AVM version 2 based on measured
	// performance. Should be able to run max hashes during stateful TEAL
	// and achieve reasonable TPS. Same opcode for different versions
	// is OK.
	{0x01, "sha256", opSHA256, proto("b:b"), 2, costly(35)},
	{0x02, "keccak256", opKeccak256, proto("b:b"), 2, costly(130)},
	{0x03, "sha512_256", opSHA512_256, proto("b:b"), 2, costly(45)},

	/*
		Tabling these changes until we offer unlimited global storage as there
		is currently a useful pattern that requires hashes on long slices to
		creating logicsigs in apps.

		{0x01, "sha256", opSHA256, proto("b:b"), unlimitedStorage, costByLength(12, 6, 8)},
		{0x02, "keccak256", opKeccak256, proto("b:b"), unlimitedStorage, costByLength(58, 4, 8)},
		{0x03, "sha512_256", opSHA512_256, proto("b:b"), 7, unlimitedStorage, costByLength(17, 5, 8)},
	*/

	{0x04, "ed25519verify", opEd25519Verify, proto("bbb:i"), 1, costly(1900).only(modeSig)},
	{0x04, "ed25519verify", opEd25519Verify, proto("bbb:i"), 5, costly(1900)},

	{0x05, "ecdsa_verify", opEcdsaVerify, proto("bbbbb:i"), 5, costByField("v", &EcdsaCurves, ecdsaVerifyCosts)},
	{0x06, "ecdsa_pk_decompress", opEcdsaPkDecompress, proto("b:bb"), 5, costByField("v", &EcdsaCurves, ecdsaDecompressCosts)},
	{0x07, "ecdsa_pk_recover", opEcdsaPkRecover, proto("bibb:bb"), 5, field("v", &EcdsaCurves).costs(2000)},

	{0x08, "+", opPlus, proto("ii:i"), 1, detDefault()},
	{0x09, "-", opMinus, proto("ii:i"), 1, detDefault()},
	{0x0a, "/", opDiv, proto("ii:i"), 1, detDefault()},
	{0x0b, "*", opMul, proto("ii:i"), 1, detDefault()},
	{0x0c, "<", opLt, proto("ii:i"), 1, detDefault()},
	{0x0d, ">", opGt, proto("ii:i"), 1, detDefault()},
	{0x0e, "<=", opLe, proto("ii:i"), 1, detDefault()},
	{0x0f, ">=", opGe, proto("ii:i"), 1, detDefault()},
	{0x10, "&&", opAnd, proto("ii:i"), 1, detDefault()},
	{0x11, "||", opOr, proto("ii:i"), 1, detDefault()},
	{0x12, "==", opEq, proto("aa:i"), 1, typed(typeEquals)},
	{0x13, "!=", opNeq, proto("aa:i"), 1, typed(typeEquals)},
	{0x14, "!", opNot, proto("i:i"), 1, detDefault()},
	{0x15, "len", opLen, proto("b:i"), 1, detDefault()},
	{0x16, "itob", opItob, proto("i:b"), 1, detDefault()},
	{0x17, "btoi", opBtoi, proto("b:i"), 1, detDefault()},
	{0x18, "%", opModulo, proto("ii:i"), 1, detDefault()},
	{0x19, "|", opBitOr, proto("ii:i"), 1, detDefault()},
	{0x1a, "&", opBitAnd, proto("ii:i"), 1, detDefault()},
	{0x1b, "^", opBitXor, proto("ii:i"), 1, detDefault()},
	{0x1c, "~", opBitNot, proto("i:i"), 1, detDefault()},
	{0x1d, "mulw", opMulw, proto("ii:ii"), 1, detDefault()},
	{0x1e, "addw", opAddw, proto("ii:ii"), 2, detDefault()},
	{0x1f, "divmodw", opDivModw, proto("iiii:iiii"), 4, costly(20)},

	{0x20, "intcblock", opIntConstBlock, proto(":"), 1, constants(asmIntCBlock, checkIntImmArgs, "uint ...", immInts)},
	{0x21, "intc", opIntConstLoad, proto(":i"), 1, immediates("i").assembler(asmIntC)},
	{0x22, "intc_0", opIntConst0, proto(":i"), 1, detDefault()},
	{0x23, "intc_1", opIntConst1, proto(":i"), 1, detDefault()},
	{0x24, "intc_2", opIntConst2, proto(":i"), 1, detDefault()},
	{0x25, "intc_3", opIntConst3, proto(":i"), 1, detDefault()},
	{0x26, "bytecblock", opByteConstBlock, proto(":"), 1, constants(asmByteCBlock, checkByteImmArgs, "bytes ...", immBytess)},
	{0x27, "bytec", opByteConstLoad, proto(":b"), 1, immediates("i").assembler(asmByteC)},
	{0x28, "bytec_0", opByteConst0, proto(":b"), 1, detDefault()},
	{0x29, "bytec_1", opByteConst1, proto(":b"), 1, detDefault()},
	{0x2a, "bytec_2", opByteConst2, proto(":b"), 1, detDefault()},
	{0x2b, "bytec_3", opByteConst3, proto(":b"), 1, detDefault()},
	{0x2c, "arg", opArg, proto(":b"), 1, immediates("n").only(modeSig).assembler(asmArg)},
	{0x2d, "arg_0", opArg0, proto(":b"), 1, only(modeSig)},
	{0x2e, "arg_1", opArg1, proto(":b"), 1, only(modeSig)},
	{0x2f, "arg_2", opArg2, proto(":b"), 1, only(modeSig)},
	{0x30, "arg_3", opArg3, proto(":b"), 1, only(modeSig)},
	// txn, gtxn, and gtxns are also implemented as pseudoOps to choose
	// between scalar and array version based on number of immediates.
	{0x31, "txn", opTxn, proto(":a"), 1, field("f", &TxnScalarFields)},
	{0x32, "global", opGlobal, proto(":a"), 1, field("f", &GlobalFields)},
	{0x33, "gtxn", opGtxn, proto(":a"), 1, immediates("t", "f").field("f", &TxnScalarFields)},
	{0x34, "load", opLoad, proto(":a"), 1, immediates("i").typed(typeLoad)},
	{0x35, "store", opStore, proto("a:"), 1, immediates("i").typed(typeStore)},
	{0x36, "txna", opTxna, proto(":a"), 2, immediates("f", "i").field("f", &TxnArrayFields)},
	{0x37, "gtxna", opGtxna, proto(":a"), 2, immediates("t", "f", "i").field("f", &TxnArrayFields)},
	// Like gtxn, but gets txn index from stack, rather than immediate arg
	{0x38, "gtxns", opGtxns, proto("i:a"), 3, immediates("f").field("f", &TxnScalarFields)},
	{0x39, "gtxnsa", opGtxnsa, proto("i:a"), 3, immediates("f", "i").field("f", &TxnArrayFields)},
	// Group scratch space access
	{0x3a, "gload", opGload, proto(":a"), 4, immediates("t", "i").only(modeApp)},
	{0x3b, "gloads", opGloads, proto("i:a"), 4, immediates("i").only(modeApp)},
	// Access creatable IDs (consider deprecating, as txn CreatedAssetID, CreatedApplicationID should be enough
	{0x3c, "gaid", opGaid, proto(":i"), 4, immediates("t").only(modeApp)},
	{0x3d, "gaids", opGaids, proto("i:i"), 4, only(modeApp)},

	// Like load/store, but scratch slot taken from TOS instead of immediate
	{0x3e, "loads", opLoads, proto("i:a"), 5, typed(typeLoads)},
	{0x3f, "stores", opStores, proto("ia:"), 5, typed(typeStores)},

	{0x40, "bnz", opBnz, proto("i:"), 1, detBranch()},
	{0x41, "bz", opBz, proto("i:"), 2, detBranch()},
	{0x42, "b", opB, proto(":"), 2, detBranch()},
	{0x43, "return", opReturn, proto("i:x"), 2, detDefault()},
	{0x44, "assert", opAssert, proto("i:"), 3, detDefault()},
	{0x45, "bury", opBury, proto("a:"), fpVersion, immediates("n").typed(typeBury)},
	{0x46, "popn", opPopN, proto(":", "[N items]", ""), fpVersion, immediates("n").typed(typePopN).trust()},
	{0x47, "dupn", opDupN, proto("a:", "", "A, [N copies of A]"), fpVersion, immediates("n").typed(typeDupN).trust()},
	{0x48, "pop", opPop, proto("a:"), 1, detDefault()},
	{0x49, "dup", opDup, proto("a:aa", "A, A"), 1, typed(typeDup)},
	{0x4a, "dup2", opDup2, proto("aa:aaaa", "A, B, A, B"), 2, typed(typeDupTwo)},
	{0x4b, "dig", opDig, proto("a:aa", "A, [N items]", "A, [N items], A"), 3, immediates("n").typed(typeDig)},
	{0x4c, "swap", opSwap, proto("aa:aa", "B, A"), 3, typed(typeSwap)},
	{0x4d, "select", opSelect, proto("aai:a", "A or B"), 3, typed(typeSelect)},
	{0x4e, "cover", opCover, proto("a:a", "[N items], A", "A, [N items]"), 5, immediates("n").typed(typeCover)},
	{0x4f, "uncover", opUncover, proto("a:a", "A, [N items]", "[N items], A"), 5, immediates("n").typed(typeUncover)},

	// byteslice processing / StringOps
	{0x50, "concat", opConcat, proto("bb:b"), 2, detDefault()},
	{0x51, "substring", opSubstring, proto("b:b"), 2, immediates("s", "e").assembler(asmSubstring)},
	{0x52, "substring3", opSubstring3, proto("bii:b"), 2, detDefault()},
	{0x53, "getbit", opGetBit, proto("ai:i"), 3, detDefault()},
	{0x54, "setbit", opSetBit, proto("aii:a"), 3, typed(typeSetBit)},
	{0x55, "getbyte", opGetByte, proto("bi:i"), 3, detDefault()},
	{0x56, "setbyte", opSetByte, proto("bii:b"), 3, detDefault()},
	{0x57, "extract", opExtract, proto("b:b"), 5, immediates("s", "l")},
	{0x58, "extract3", opExtract3, proto("bii:b"), 5, detDefault()},
	{0x59, "extract_uint16", opExtract16Bits, proto("bi:i"), 5, detDefault()},
	{0x5a, "extract_uint32", opExtract32Bits, proto("bi:i"), 5, detDefault()},
	{0x5b, "extract_uint64", opExtract64Bits, proto("bi:i"), 5, detDefault()},
	{0x5c, "replace2", opReplace2, proto("bb:b"), 7, immediates("s")},
	{0x5d, "replace3", opReplace3, proto("bib:b"), 7, detDefault()},
	{0x5e, "base64_decode", opBase64Decode, proto("b:b"), fidoVersion, field("e", &Base64Encodings).costByLength(1, 1, 16, 0)},
	{0x5f, "json_ref", opJSONRef, proto("bb:a"), fidoVersion, field("r", &JSONRefTypes).costByLength(25, 2, 7, 1)},

	{0x60, "balance", opBalance, proto("i:i"), 2, only(modeApp)},
	{0x60, "balance", opBalance, proto("a:i"), directRefEnabledVersion, only(modeApp)},
	{0x61, "app_opted_in", opAppOptedIn, proto("ii:i"), 2, only(modeApp)},
	{0x61, "app_opted_in", opAppOptedIn, proto("ai:i"), directRefEnabledVersion, only(modeApp)},
	{0x62, "app_local_get", opAppLocalGet, proto("ib:a"), 2, only(modeApp)},
	{0x62, "app_local_get", opAppLocalGet, proto("ab:a"), directRefEnabledVersion, only(modeApp)},
	{0x63, "app_local_get_ex", opAppLocalGetEx, proto("iib:ai"), 2, only(modeApp)},
	{0x63, "app_local_get_ex", opAppLocalGetEx, proto("aib:ai"), directRefEnabledVersion, only(modeApp)},
	{0x64, "app_global_get", opAppGlobalGet, proto("b:a"), 2, only(modeApp)},
	{0x65, "app_global_get_ex", opAppGlobalGetEx, proto("ib:ai"), 2, only(modeApp)},
	{0x66, "app_local_put", opAppLocalPut, proto("iba:"), 2, only(modeApp)},
	{0x66, "app_local_put", opAppLocalPut, proto("aba:"), directRefEnabledVersion, only(modeApp)},
	{0x67, "app_global_put", opAppGlobalPut, proto("ba:"), 2, only(modeApp)},
	{0x68, "app_local_del", opAppLocalDel, proto("ib:"), 2, only(modeApp)},
	{0x68, "app_local_del", opAppLocalDel, proto("ab:"), directRefEnabledVersion, only(modeApp)},
	{0x69, "app_global_del", opAppGlobalDel, proto("b:"), 2, only(modeApp)},

	{0x70, "asset_holding_get", opAssetHoldingGet, proto("ii:ai"), 2, field("f", &AssetHoldingFields).only(modeApp)},
	{0x70, "asset_holding_get", opAssetHoldingGet, proto("ai:ai"), directRefEnabledVersion, field("f", &AssetHoldingFields).only(modeApp)},
	{0x71, "asset_params_get", opAssetParamsGet, proto("i:ai"), 2, field("f", &AssetParamsFields).only(modeApp)},
	{0x72, "app_params_get", opAppParamsGet, proto("i:ai"), 5, field("f", &AppParamsFields).only(modeApp)},
	{0x73, "acct_params_get", opAcctParamsGet, proto("a:ai"), 6, field("f", &AcctParamsFields).only(modeApp)},

	{0x78, "min_balance", opMinBalance, proto("i:i"), 3, only(modeApp)},
	{0x78, "min_balance", opMinBalance, proto("a:i"), directRefEnabledVersion, only(modeApp)},

	// Immediate bytes and ints. Smaller code size for single use of constant.
	{0x80, "pushbytes", opPushBytes, proto(":b"), 3, constants(asmPushBytes, opPushBytes, "bytes", immBytes)},
	{0x81, "pushint", opPushInt, proto(":i"), 3, constants(asmPushInt, opPushInt, "uint", immInt)},
	{0x82, "pushbytess", opPushBytess, proto(":", "", "[N items]"), 8, constants(asmPushBytess, checkByteImmArgs, "bytes ...", immBytess).typed(typePushBytess).trust()},
	{0x83, "pushints", opPushInts, proto(":", "", "[N items]"), 8, constants(asmPushInts, checkIntImmArgs, "uint ...", immInts).typed(typePushInts).trust()},

	{0x84, "ed25519verify_bare", opEd25519VerifyBare, proto("bbb:i"), 7, costly(1900)},

	// "Function oriented"
	{0x88, "callsub", opCallSub, proto(":"), 4, detBranch()},
	{0x89, "retsub", opRetSub, proto(":"), 4, detDefault().trust()},
	// protoByte is a named constant because opCallSub needs to know it.
	{protoByte, "proto", opProto, proto(":"), fpVersion, immediates("a", "r").typed(typeProto)},
	{0x8b, "frame_dig", opFrameDig, proto(":a"), fpVersion, immKinded(immInt8, "i").typed(typeFrameDig)},
	{0x8c, "frame_bury", opFrameBury, proto("a:"), fpVersion, immKinded(immInt8, "i").typed(typeFrameBury)},
	{0x8d, "switch", opSwitch, proto("i:"), 8, detSwitch()},
	{0x8e, "match", opMatch, proto(":", "[A1, A2, ..., AN], B", ""), 8, detSwitch().trust()},

	// More math
	{0x90, "shl", opShiftLeft, proto("ii:i"), 4, detDefault()},
	{0x91, "shr", opShiftRight, proto("ii:i"), 4, detDefault()},
	{0x92, "sqrt", opSqrt, proto("i:i"), 4, costly(4)},
	{0x93, "bitlen", opBitLen, proto("a:i"), 4, detDefault()},
	{0x94, "exp", opExp, proto("ii:i"), 4, detDefault()},
	{0x95, "expw", opExpw, proto("ii:ii"), 4, costly(10)},
	{0x96, "bsqrt", opBytesSqrt, proto("b:b"), 6, costly(40)},
	{0x97, "divw", opDivw, proto("iii:i"), 6, detDefault()},
	{0x98, "sha3_256", opSHA3_256, proto("b:b"), 7, costly(130)},
	/* Will end up following keccak256 -
	{0x98, "sha3_256", opSHA3_256, proto("b:b"), unlimitedStorage, costByLength(58, 4, 8)},},
	*/

	{0x99, "bn256_add", opBn256Add, proto("bb:b"), pairingVersion, costly(70)},
	{0x9a, "bn256_scalar_mul", opBn256ScalarMul, proto("bb:b"), pairingVersion, costly(970)},
	{0x9b, "bn256_pairing", opBn256Pairing, proto("bb:i"), pairingVersion, costly(8700)},

	// Byteslice math.
	{0xa0, "b+", opBytesPlus, proto("bb:b"), 4, costly(10)},
	{0xa1, "b-", opBytesMinus, proto("bb:b"), 4, costly(10)},
	{0xa2, "b/", opBytesDiv, proto("bb:b"), 4, costly(20)},
	{0xa3, "b*", opBytesMul, proto("bb:b"), 4, costly(20)},
	{0xa4, "b<", opBytesLt, proto("bb:i"), 4, detDefault()},
	{0xa5, "b>", opBytesGt, proto("bb:i"), 4, detDefault()},
	{0xa6, "b<=", opBytesLe, proto("bb:i"), 4, detDefault()},
	{0xa7, "b>=", opBytesGe, proto("bb:i"), 4, detDefault()},
	{0xa8, "b==", opBytesEq, proto("bb:i"), 4, detDefault()},
	{0xa9, "b!=", opBytesNeq, proto("bb:i"), 4, detDefault()},
	{0xaa, "b%", opBytesModulo, proto("bb:b"), 4, costly(20)},
	{0xab, "b|", opBytesBitOr, proto("bb:b"), 4, costly(6)},
	{0xac, "b&", opBytesBitAnd, proto("bb:b"), 4, costly(6)},
	{0xad, "b^", opBytesBitXor, proto("bb:b"), 4, costly(6)},
	{0xae, "b~", opBytesBitNot, proto("b:b"), 4, costly(4)},
	{0xaf, "bzero", opBytesZero, proto("i:b"), 4, detDefault()},

	// AVM "effects"
	{0xb0, "log", opLog, proto("b:"), 5, only(modeApp)},
	{0xb1, "itxn_begin", opTxBegin, proto(":"), 5, only(modeApp)},
	{0xb2, "itxn_field", opItxnField, proto("a:"), 5, immediates("f").typed(typeTxField).field("f", &TxnFields).only(modeApp).assembler(asmItxnField)},
	{0xb3, "itxn_submit", opItxnSubmit, proto(":"), 5, only(modeApp)},
	{0xb4, "itxn", opItxn, proto(":a"), 5, field("f", &TxnScalarFields).only(modeApp).assembler(asmItxn)},
	{0xb5, "itxna", opItxna, proto(":a"), 5, immediates("f", "i").field("f", &TxnArrayFields).only(modeApp)},
	{0xb6, "itxn_next", opItxnNext, proto(":"), 6, only(modeApp)},
	{0xb7, "gitxn", opGitxn, proto(":a"), 6, immediates("t", "f").field("f", &TxnFields).only(modeApp).assembler(asmGitxn)},
	{0xb8, "gitxna", opGitxna, proto(":a"), 6, immediates("t", "f", "i").field("f", &TxnArrayFields).only(modeApp)},

	// Unlimited Global Storage - Boxes
	{0xb9, "box_create", opBoxCreate, proto("bi:i"), boxVersion, only(modeApp)},
	{0xba, "box_extract", opBoxExtract, proto("bii:b"), boxVersion, only(modeApp)},
	{0xbb, "box_replace", opBoxReplace, proto("bib:"), boxVersion, only(modeApp)},
	{0xbc, "box_del", opBoxDel, proto("b:i"), boxVersion, only(modeApp)},
	{0xbd, "box_len", opBoxLen, proto("b:ii"), boxVersion, only(modeApp)},
	{0xbe, "box_get", opBoxGet, proto("b:bi"), boxVersion, only(modeApp)},
	{0xbf, "box_put", opBoxPut, proto("bb:"), boxVersion, only(modeApp)},

	// Dynamic indexing
	{0xc0, "txnas", opTxnas, proto("i:a"), 5, field("f", &TxnArrayFields)},
	{0xc1, "gtxnas", opGtxnas, proto("i:a"), 5, immediates("t", "f").field("f", &TxnArrayFields)},
	{0xc2, "gtxnsas", opGtxnsas, proto("ii:a"), 5, field("f", &TxnArrayFields)},
	{0xc3, "args", opArgs, proto("i:b"), 5, only(modeSig)},
	{0xc4, "gloadss", opGloadss, proto("ii:a"), 6, only(modeApp)},
	{0xc5, "itxnas", opItxnas, proto("i:a"), 6, field("f", &TxnArrayFields).only(modeApp)},
	{0xc6, "gitxnas", opGitxnas, proto("i:a"), 6, immediates("t", "f").field("f", &TxnArrayFields).only(modeApp)},

	// randomness support
	{0xd0, "vrf_verify", opVrfVerify, proto("bbb:bi"), randomnessVersion, field("s", &VrfStandards).costs(5700)},
	{0xd1, "block", opBlock, proto("i:a"), randomnessVersion, field("f", &BlockFields)},
}

func parseBinaryArgs(args []string) (val []byte, consumed int, err error) {
	arg := args[0]
	if strings.HasPrefix(arg, "base32(") || strings.HasPrefix(arg, "b32(") {
		open := strings.IndexRune(arg, '(')
		close := strings.IndexRune(arg, ')')
		if close == -1 {
			err = errors.New("byte base32 arg lacks close paren")
			return
		}
		val, err = base32DecodeAnyPadding(arg[open+1 : close])
		if err != nil {
			return
		}
		consumed = 1
	} else if strings.HasPrefix(arg, "base64(") || strings.HasPrefix(arg, "b64(") {
		open := strings.IndexRune(arg, '(')
		close := strings.IndexRune(arg, ')')
		if close == -1 {
			err = errors.New("byte base64 arg lacks close paren")
			return
		}
		val, err = base64.StdEncoding.DecodeString(arg[open+1 : close])
		if err != nil {
			return
		}
		consumed = 1
	} else if strings.HasPrefix(arg, "0x") {
		val, err = hex.DecodeString(arg[2:])
		if err != nil {
			return
		}
		consumed = 1
	} else if arg == "base32" || arg == "b32" {
		if len(args) < 2 {
			err = fmt.Errorf("need literal after 'byte %s'", arg)
			return
		}
		val, err = base32DecodeAnyPadding(args[1])
		if err != nil {
			return
		}
		consumed = 2
	} else if arg == "base64" || arg == "b64" {
		if len(args) < 2 {
			err = fmt.Errorf("need literal after 'byte %s'", arg)
			return
		}
		val, err = base64.StdEncoding.DecodeString(args[1])
		if err != nil {
			return
		}
		consumed = 2
	} else if len(arg) > 1 && arg[0] == '"' && arg[len(arg)-1] == '"' {
		val, err = parseStringLiteral(arg)
		consumed = 1
	} else {
		err = fmt.Errorf("byte arg did not parse: %v", arg)
		return
	}
	return
}

func parseStringLiteral(input string) (result []byte, err error) {
	start := 0
	end := len(input) - 1
	if input[start] != '"' || input[end] != '"' {
		return nil, fmt.Errorf("no quotes")
	}
	start++

	escapeSeq := false
	hexSeq := false
	result = make([]byte, 0, end-start+1)

	// skip first and last quotes
	pos := start
	for pos < end {
		char := input[pos]
		if char == '\\' && !escapeSeq {
			if hexSeq {
				return nil, fmt.Errorf("escape seq inside hex number")
			}
			escapeSeq = true
			pos++
			continue
		}
		if escapeSeq {
			escapeSeq = false
			switch char {
			case 'n':
				char = '\n'
			case 'r':
				char = '\r'
			case 't':
				char = '\t'
			case '\\':
				char = '\\'
			case '"':
				char = '"'
			case 'x':
				hexSeq = true
				pos++
				continue
			default:
				return nil, fmt.Errorf("invalid escape seq \\%c", char)
			}
		}
		if hexSeq {
			hexSeq = false
			if pos >= len(input)-2 { // count a closing quote
				return nil, fmt.Errorf("non-terminated hex seq")
			}
			num, err := strconv.ParseUint(input[pos:pos+2], 16, 8)
			if err != nil {
				return nil, err
			}
			char = uint8(num)
			pos++
		}

		result = append(result, char)
		pos++
	}
	if escapeSeq || hexSeq {
		return nil, fmt.Errorf("non-terminated escape seq")
	}

	return
}

func base32DecodeAnyPadding(x string) (val []byte, err error) {
	val, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(x)
	if err != nil {
		// try again with standard padding
		var e2 error
		val, e2 = base32.StdEncoding.DecodeString(x)
		if e2 == nil {
			err = nil
		}
	}
	return
}
