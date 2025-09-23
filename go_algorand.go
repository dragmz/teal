package teal

import (
	"encoding/base32"
	"fmt"
	"strconv"

	"github.com/dragmz/teal/internal/protocol"
	"github.com/dragmz/teal/internal/transactions"
)

// Copyright (C) 2019-2022 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

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

func addExtra(original string, extra string) string {
	if len(original) == 0 {
		return extra
	}
	if len(extra) == 0 {
		return original
	}
	sep := ". "
	if original[len(original)-1] == '.' {
		sep = " "
	}
	return original + sep + extra
}

// short description of every op
var opDocByName = map[string]string{
	"err":                 "Fail immediately.",
	"sha256":              "SHA256 hash of value A, yields [32]byte",
	"keccak256":           "Keccak256 hash of value A, yields [32]byte",
	"sha512_256":          "SHA512_256 hash of value A, yields [32]byte",
	"sha3_256":            "SHA3_256 hash of value A, yields [32]byte",
	"sumhash512":          "sumhash512 of value A, yields [64]byte",
	"falcon_verify":       "for (data A, compressed-format signature B, pubkey C) verify the signature of data against the pubke",
	"ed25519verify":       "for (data A, signature B, pubkey C) verify the signature of (\"ProgData\" || program_hash || data) against the pubkey => {0 or 1}",
	"ed25519verify_bare":  "for (data A, signature B, pubkey C) verify the signature of the data against the pubkey => {0 or 1}",
	"ecdsa_verify":        "for (data A, signature B, C and pubkey D, E) verify the signature of the data against the pubkey => {0 or 1}",
	"ecdsa_pk_decompress": "decompress pubkey A into components X, Y",
	"ecdsa_pk_recover":    "for (data A, recovery id B, signature C, D) recover a public key",

	"ec_add":              "for curve points A and B, return the curve point A + B",
	"ec_scalar_mul":       "for curve point A and scalar B, return the curve point BA, the point A multiplied by the scalar B.",
	"ec_pairing_check":    "1 if the product of the pairing of each point in A with its respective point in B is equal to the identity element of the target group Gt, else 0",
	"ec_multi_scalar_mul": "for curve points A and scalars B, return curve point B0A0 + B1A1 + B2A2 + ... + BnAn",
	"ec_subgroup_check":   "1 if A is in the main prime-order subgroup of G (including the point at infinity) else 0. Program fails if A is not in G at all.",
	"ec_map_to":           "maps field element A to group G",

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
	"box_splice":  "set box A to contain its previous bytes up to index B, followed by D, followed by the original bytes of A that began at index B+C.",
	"box_del":     "delete box named A if it exists. Return 1 if A existed, 0 otherwise",
	"box_len":     "X is the length of box A if A exists, else 0. Y is 1 if A exists, else 0.",
	"box_get":     "X is the contents of box A if A exists, else ''. Y is 1 if A exists, else 0.",
	"box_put":     "replaces the contents of box A with byte-array B. Fails if A exists and len(B) != len(box A). Creates A if it does not exist",
	"box_resize":  "change the size of box named A to be of length B, adding zero bytes to end or removing bytes from the end, as needed. Fail if the name A is empty, A is not an existing box, or B exceeds 32,768.",
}

var opDocExtras = map[string]string{
	"vrf_verify":          "`VrfAlgorand` is the VRF used in Algorand. It is ECVRF-ED25519-SHA512-Elligator2, specified in the IETF internet draft [draft-irtf-cfrg-vrf-03](https://datatracker.ietf.org/doc/draft-irtf-cfrg-vrf/03/).",
	"ed25519verify":       "The 32 byte public key is the last element on the stack, preceded by the 64 byte signature at the second-to-last element on the stack, preceded by the data which was signed at the third-to-last element on the stack.",
	"ecdsa_verify":        "The 32 byte Y-component of a public key is the last element on the stack, preceded by X-component of a pubkey, preceded by S and R components of a signature, preceded by the data that is fifth element on the stack. All values are big-endian encoded. The signed data must be 32 bytes long, and signatures in lower-S form are only accepted.",
	"ecdsa_pk_decompress": "The 33 byte public key in a compressed form to be decompressed into X and Y (top) components. All values are big-endian encoded.",
	"ecdsa_pk_recover":    "S (top) and R elements of a signature, recovery id and data (bottom) are expected on the stack and used to deriver a public key. All values are big-endian encoded. The signed data must be 32 bytes long.",

	"ec_add": "A and B are curve points in affine representation: field element X concatenated with field element Y. " +
		"Field element `Z` is encoded as follows.\n" +
		"For the base field elements (Fp), `Z` is encoded as a big-endian number and must be lower than the field modulus.\n" +
		"For the quadratic field extension (Fp2), `Z` is encoded as the concatenation of the individual encoding of the coefficients. " +
		"For an Fp2 element of the form `Z = Z0 + Z1 i`, where `i` is a formal quadratic non-residue, the encoding of Z is the concatenation of the encoding of `Z0` and `Z1` in this order. (`Z0` and `Z1` must be less than the field modulus).\n\n" +
		"The point at infinity is encoded as `(X,Y) = (0,0)`.\n" +
		"Groups G1 and G2 are denoted additively.\n\n" +
		"Fails if A or B is not in G.\n" +
		"A and/or B are allowed to be the point at infinity.\n" +
		"Does _not_ check if A and B are in the main prime-order subgroup.",

	"ec_scalar_mul":       "A is a curve point encoded and checked as described in `ec_add`. Scalar B is interpreted as a big-endian unsigned integer. Fails if B exceeds 32 bytes.",
	"ec_pairing_check":    "A and B are concatenated points, encoded and checked as described in `ec_add`. A contains points of the group G, B contains points of the associated group (G2 if G is G1, and vice versa). Fails if A and B have a different number of points, or if any point is not in its described group or outside the main prime-order subgroup - a stronger condition than other opcodes. AVM values are limited to 4096 bytes, so `ec_pairing_check` is limited by the size of the points in the groups being operated upon.",
	"ec_multi_scalar_mul": "A is a list of concatenated points, encoded and checked as described in `ec_add`. B is a list of concatenated scalars which, unlike ec_scalar_mul, must all be exactly 32 bytes long.\nThe name `ec_multi_scalar_mul` was chosen to reflect common usage, but a more consistent name would be `ec_multi_scalar_mul`. AVM values are limited to 4096 bytes, so `ec_multi_scalar_mul` is limited by the size of the points in the group being operated upon.",
	"ec_map_to": "BN254 points are mapped by the SVDW map. BLS12-381 points are mapped by the SSWU map.\n" +
		"G1 element inputs are base field elements and G2 element inputs are quadratic field elements, with nearly the same encoding rules (for field elements) as defined in `ec_add`. There is one difference of encoding rule: G1 element inputs do not need to be 0-padded if they fit in less than 32 bytes for BN254 and less than 48 bytes for BLS12-381. (As usual, the empty byte array represents 0.) G2 elements inputs need to be always have the required size.",

	"bnz":               "The `bnz` instruction opcode 0x40 is followed by two immediate data bytes which are a high byte first and low byte second which together form a 16 bit offset which the instruction may branch to. For a bnz instruction at `pc`, if the last element of the stack is not zero then branch to instruction at `pc + 3 + N`, else proceed to next instruction at `pc + 3`. Branch targets must be aligned instructions. (e.g. Branching to the second byte of a 2 byte op will be rejected.) Starting at v4, the offset is treated as a signed 16 bit integer allowing for backward branches and looping. In prior version (v1 to v3), branch offsets are limited to forward branches only, 0-0x7fff.\n\nAt v2 it became allowed to branch to the end of the program exactly after the last instruction: bnz to byte N (with 0-indexing) was illegal for a TEAL program with N bytes before v2, and is legal after it. This change eliminates the need for a last instruction of no-op as a branch target at the end. (Branching beyond the end--in other words, to a byte larger than N--is still illegal and will cause the program to fail.)",
	"bz":                "See `bnz` for details on how branches work. `bz` inverts the behavior of `bnz`.",
	"b":                 "See `bnz` for details on how branches work. `b` always jumps to the offset.",
	"callsub":           "The call stack is separate from the data stack. Only `callsub`, `retsub`, and `proto` manipulate it.",
	"proto":             "Fails unless the last instruction executed was a `callsub`.",
	"retsub":            "If the current frame was prepared by `proto A R`, `retsub` will remove the 'A' arguments from the stack, move the `R` return values down, and pop any stack locations above the relocated return values.",
	"intcblock":         "`intcblock` loads following program bytes into an array of integer constants in the evaluator. These integer constants can be referred to by `intc` and `intc_*` which will push the value onto the stack. Subsequent calls to `intcblock` reset and replace the integer constants available to the script.",
	"bytecblock":        "`bytecblock` loads the following program bytes into an array of byte-array constants in the evaluator. These constants can be referred to by `bytec` and `bytec_*` which will push the value onto the stack. Subsequent calls to `bytecblock` reset and replace the bytes constants available to the script.",
	"*":                 "Overflow is an error condition which halts execution and fails the transaction. Full precision is available from `mulw`.",
	"+":                 "Overflow is an error condition which halts execution and fails the transaction. Full precision is available from `addw`.",
	"/":                 "`divmodw` is available to divide the two-element values produced by `mulw` and `addw`.",
	"bitlen":            "bitlen interprets arrays as big-endian integers, unlike setbit/getbit",
	"divw":              "The notation A,B indicates that A and B are interpreted as a uint128 value, with A as the high uint64 and B the low.",
	"divmodw":           "The notation J,K indicates that two uint64 values J and K are interpreted as a uint128 value, with J as the high uint64 and K the low.",
	"gtxn":              "for notes on transaction fields available, see `txn`. If this transaction is _i_ in the group, `gtxn i field` is equivalent to `txn field`.",
	"gtxns":             "for notes on transaction fields available, see `txn`. If top of stack is _i_, `gtxns field` is equivalent to `gtxn _i_ field`. gtxns exists so that _i_ can be calculated, often based on the index of the current transaction.",
	"gload":             "`gload` fails unless the requested transaction is an ApplicationCall and T < GroupIndex.",
	"gloads":            "`gloads` fails unless the requested transaction is an ApplicationCall and A < GroupIndex.",
	"gaid":              "`gaid` fails unless the requested transaction created an asset or application and T < GroupIndex.",
	"gaids":             "`gaids` fails unless the requested transaction created an asset or application and A < GroupIndex.",
	"btoi":              "`btoi` fails if the input is longer than 8 bytes.",
	"concat":            "`concat` fails if the result would be greater than 4096 bytes.",
	"pushbytes":         "pushbytes args are not added to the bytecblock during assembly processes",
	"pushbytess":        "pushbytess args are not added to the bytecblock during assembly processes",
	"pushint":           "pushint args are not added to the intcblock during assembly processes",
	"pushints":          "pushints args are not added to the intcblock during assembly processes",
	"getbit":            "see explanation of bit ordering in setbit",
	"setbit":            "When A is a uint64, index 0 is the least significant bit. Setting bit 3 to 1 on the integer 0 yields 8, or 2^3. When A is a byte array, index 0 is the leftmost bit of the leftmost byte. Setting bits 0 through 11 to 1 in a 4-byte-array of 0s yields the byte array 0xfff00000. Setting bit 3 to 1 on the 1-byte-array 0x00 yields the byte array 0x10.",
	"balance":           "params: Txn.Accounts offset (or, since v4, an _available_ account address), _available_ application id (or, since v4, a Txn.ForeignApps offset). Return: value.",
	"min_balance":       "params: Txn.Accounts offset (or, since v4, an _available_ account address), _available_ application id (or, since v4, a Txn.ForeignApps offset). Return: value.",
	"app_opted_in":      "params: Txn.Accounts offset (or, since v4, an _available_ account address), _available_ application id (or, since v4, a Txn.ForeignApps offset). Return: 1 if opted in and 0 otherwise.",
	"app_local_get":     "params: Txn.Accounts offset (or, since v4, an _available_ account address), state key. Return: value. The value is zero (of type uint64) if the key does not exist.",
	"app_local_get_ex":  "params: Txn.Accounts offset (or, since v4, an _available_ account address), _available_ application id (or, since v4, a Txn.ForeignApps offset), state key. Return: did_exist flag (top of the stack, 1 if the application and key existed and 0 otherwise), value. The value is zero (of type uint64) if the key does not exist.",
	"app_global_get_ex": "params: Txn.ForeignApps offset (or, since v4, an _available_ application id), state key. Return: did_exist flag (top of the stack, 1 if the application and key existed and 0 otherwise), value. The value is zero (of type uint64) if the key does not exist.",
	"app_global_get":    "params: state key. Return: value. The value is zero (of type uint64) if the key does not exist.",
	"app_local_put":     "params: Txn.Accounts offset (or, since v4, an _available_ account address), state key, value.",
	"app_local_del":     "params: Txn.Accounts offset (or, since v4, an _available_ account address), state key.\n\nDeleting a key which is already absent has no effect on the application local state. (In particular, it does _not_ cause the program to fail.)",
	"app_global_del":    "params: state key.\n\nDeleting a key which is already absent has no effect on the application global state. (In particular, it does _not_ cause the program to fail.)",
	"asset_holding_get": "params: Txn.Accounts offset (or, since v4, an _available_ address), asset id (or, since v4, a Txn.ForeignAssets offset). Return: did_exist flag (1 if the asset existed and 0 otherwise), value.",
	"asset_params_get":  "params: Txn.ForeignAssets offset (or, since v4, an _available_ asset id. Return: did_exist flag (1 if the asset existed and 0 otherwise), value.",
	"app_params_get":    "params: Txn.ForeignApps offset or an _available_ app id. Return: did_exist flag (1 if the application existed and 0 otherwise), value.",
	"log":               "`log` fails if called more than MaxLogCalls times in a program, or if the sum of logged bytes exceeds 1024 bytes.",
	"itxn_begin":        "`itxn_begin` initializes Sender to the application address; Fee to the minimum allowable, taking into account MinTxnFee and credit from overpaying in earlier transactions; FirstValid/LastValid to the values in the invoking transaction, and all other fields to zero or empty values.",
	"itxn_next":         "`itxn_next` initializes the transaction exactly as `itxn_begin` does",
	"itxn_field":        "`itxn_field` fails if A is of the wrong type for F, including a byte array of the wrong size for use as an address when F is an address field. `itxn_field` also fails if A is an account, asset, or app that is not _available_, or an attempt is made extend an array field beyond the limit imposed by consensus parameters. (Addresses set into asset params of acfg transactions need not be _available_.)",
	"itxn_submit":       "`itxn_submit` resets the current transaction so that it can not be resubmitted. A new `itxn_begin` is required to prepare another inner transaction.",

	"base64_decode": "*Warning*: Usage should be restricted to very rare use cases. In almost all cases, smart contracts should directly handle non-encoded byte-strings.	This opcode should only be used in cases where base64 is the only available option, e.g. interoperability with a third-party that only signs base64 strings.\n\n Decodes A using the base64 encoding E. Specify the encoding with an immediate arg either as URL and Filename Safe (`URLEncoding`) or Standard (`StdEncoding`). See [RFC 4648 sections 4 and 5](https://rfc-editor.org/rfc/rfc4648.html#section-4). It is assumed that the encoding ends with the exact number of `=` padding characters as required by the RFC. When padding occurs, any unused pad bits in the encoding must be set to zero or the decoding will fail. The special cases of `\\n` and `\\r` are allowed but completely ignored. An error will result when attempting to decode a string with a character that is not in the encoding alphabet or not one of `=`, `\\r`, or `\\n`.",
	"json_ref":      "*Warning*: Usage should be restricted to very rare use cases, as JSON decoding is expensive and quite limited. In addition, JSON objects are large and not optimized for size.\n\nAlmost all smart contracts should use simpler and smaller methods (such as the [ABI](https://arc.algorand.foundation/ARCs/arc-0004). This opcode should only be used in cases where JSON is only available option, e.g. when a third-party only signs JSON.",

	"match": "`match` consumes N+1 values from the stack. Let the top stack value be B. The following N values represent an ordered list of match cases/constants (A), where the first value (A[0]) is the deepest in the stack. The immediate arguments are an ordered list of N labels (T). `match` will branch to target T[I], where A[I] = B. If there are no matches then execution continues on to the next instruction.",

	"box_create": "Newly created boxes are filled with 0 bytes. `box_create` will fail if the referenced box already exists with a different size. Otherwise, existing boxes are unchanged by `box_create`.",
	"box_get":    "For boxes that exceed 4,096 bytes, consider `box_create`, `box_extract`, and `box_replace`",
	"box_put":    "For boxes that exceed 4,096 bytes, consider `box_create`, `box_extract`, and `box_replace`",
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

// fields

type StackType byte

const (
	// StackNone in an OpSpec shows that the op pops or yields nothing
	StackNone StackType = iota

	// StackAny in an OpSpec shows that the op pops or yield any type
	StackAny

	// StackUint64 in an OpSpec shows that the op pops or yields a uint64
	StackUint64

	// StackBytes in an OpSpec shows that the op pops or yields a []byte
	StackBytes

	StackAddress

	StackBytes32

	StackBoolean
)

// StackTypes is an alias for a list of StackType with syntactic sugar
type StackTypes []StackType

type RunMode uint64

const (
	// ModeSig is LogicSig execution
	ModeSig RunMode = 1 << iota

	// ModeApp is application/contract execution
	ModeApp

	// local constant, run in any mode
	ModeAny = ModeSig | ModeApp
)

// Copyright (C) 2019-2023 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

//go:generate stringer -type=TxnField,GlobalField,AssetParamsField,AppParamsField,AcctParamsField,AssetHoldingField,OnCompletionConstType,EcdsaCurve,EcGroup,Base64Encoding,JSONRefType,VrfStandard,BlockField -output=fields_string.go

// FieldSpec unifies the various specs for assembly, disassembly, and doc generation.
type FieldSpec interface {
	Field() byte
	Type() StackType
	OpVersion() uint64
	Note() string
	Version() uint64
}

// fieldSpecMap is something that yields a FieldSpec, given a name for the field
type fieldSpecMap interface {
	get(name string) (FieldSpec, bool)
}

// FieldGroup binds all the info for a field (names, int value, spec access) so
// they can be attached to opcodes and used by doc generation
type FieldGroup struct {
	Name  string
	Doc   string
	Names []string
	specs fieldSpecMap
}

// SpecByName returns a FieldsSpec for a name, respecting the "sparseness" of
// the Names array to hide some names
func (fg *FieldGroup) SpecByName(name string) (FieldSpec, bool) {
	if fs, ok := fg.specs.get(name); ok {
		if fg.Names[fs.Field()] != "" {
			return fs, true
		}
	}
	return nil, false
}

// TxnField is an enum type for `txn` and `gtxn`
type TxnField int

const (
	// Sender Transaction.Sender
	Sender TxnField = iota
	// Fee Transaction.Fee
	Fee
	// FirstValid Transaction.FirstValid
	FirstValid
	// FirstValidTime timestamp of block(FirstValid-1)
	FirstValidTime
	// LastValid Transaction.LastValid
	LastValid
	// Note Transaction.Note
	Note
	// Lease Transaction.Lease
	Lease
	// Receiver Transaction.Receiver
	Receiver
	// Amount Transaction.Amount
	Amount
	// CloseRemainderTo Transaction.CloseRemainderTo
	CloseRemainderTo
	// VotePK Transaction.VotePK
	VotePK
	// SelectionPK Transaction.SelectionPK
	SelectionPK
	// VoteFirst Transaction.VoteFirst
	VoteFirst
	// VoteLast Transaction.VoteLast
	VoteLast
	// VoteKeyDilution Transaction.VoteKeyDilution
	VoteKeyDilution
	// Type Transaction.Type
	Type
	// TypeEnum int(Transaction.Type)
	TypeEnum
	// XferAsset Transaction.XferAsset
	XferAsset
	// AssetAmount Transaction.AssetAmount
	AssetAmount
	// AssetSender Transaction.AssetSender
	AssetSender
	// AssetReceiver Transaction.AssetReceiver
	AssetReceiver
	// AssetCloseTo Transaction.AssetCloseTo
	AssetCloseTo
	// GroupIndex i for txngroup[i] == Txn
	GroupIndex
	// TxID Transaction.ID()
	TxID
	// ApplicationID basics.AppIndex
	ApplicationID
	// OnCompletion OnCompletion
	OnCompletion
	// ApplicationArgs  [][]byte
	ApplicationArgs
	// NumAppArgs len(ApplicationArgs)
	NumAppArgs
	// Accounts []basics.Address
	Accounts
	// NumAccounts len(Accounts)
	NumAccounts
	// ApprovalProgram []byte
	ApprovalProgram
	// ClearStateProgram []byte
	ClearStateProgram
	// RekeyTo basics.Address
	RekeyTo
	// ConfigAsset basics.AssetIndex
	ConfigAsset
	// ConfigAssetTotal AssetParams.Total
	ConfigAssetTotal
	// ConfigAssetDecimals AssetParams.Decimals
	ConfigAssetDecimals
	// ConfigAssetDefaultFrozen AssetParams.AssetDefaultFrozen
	ConfigAssetDefaultFrozen
	// ConfigAssetUnitName AssetParams.UnitName
	ConfigAssetUnitName
	// ConfigAssetName AssetParams.AssetName
	ConfigAssetName
	// ConfigAssetURL AssetParams.URL
	ConfigAssetURL
	// ConfigAssetMetadataHash AssetParams.MetadataHash
	ConfigAssetMetadataHash
	// ConfigAssetManager AssetParams.Manager
	ConfigAssetManager
	// ConfigAssetReserve AssetParams.Reserve
	ConfigAssetReserve
	// ConfigAssetFreeze AssetParams.Freeze
	ConfigAssetFreeze
	// ConfigAssetClawback AssetParams.Clawback
	ConfigAssetClawback
	//FreezeAsset  basics.AssetIndex
	FreezeAsset
	// FreezeAssetAccount basics.Address
	FreezeAssetAccount
	// FreezeAssetFrozen bool
	FreezeAssetFrozen
	// Assets []basics.AssetIndex
	Assets
	// NumAssets len(ForeignAssets)
	NumAssets
	// Applications []basics.AppIndex
	Applications
	// NumApplications len(ForeignApps)
	NumApplications

	// GlobalNumUint uint64
	GlobalNumUint
	// GlobalNumByteSlice uint64
	GlobalNumByteSlice
	// LocalNumUint uint64
	LocalNumUint
	// LocalNumByteSlice uint64
	LocalNumByteSlice

	// ExtraProgramPages AppParams.ExtraProgramPages
	ExtraProgramPages

	// Nonparticipation Transaction.Nonparticipation
	Nonparticipation

	// Logs Transaction.ApplyData.EvalDelta.Logs
	Logs

	// NumLogs len(Logs)
	NumLogs

	// CreatedAssetID Transaction.ApplyData.EvalDelta.ConfigAsset
	CreatedAssetID

	// CreatedApplicationID Transaction.ApplyData.EvalDelta.ApplicationID
	CreatedApplicationID

	// LastLog Logs[len(Logs)-1]
	LastLog

	// StateProofPK Transaction.StateProofPK
	StateProofPK

	// ApprovalProgramPages [][]byte
	ApprovalProgramPages

	// NumApprovalProgramPages = len(ApprovalProgramPages) // 4096
	NumApprovalProgramPages

	// ClearStateProgramPages [][]byte
	ClearStateProgramPages

	// NumClearStateProgramPages = len(ClearStateProgramPages) // 4096
	NumClearStateProgramPages

	invalidTxnField // compile-time constant for number of fields
)

func txnFieldSpecByField(f TxnField) (txnFieldSpec, bool) {
	if int(f) >= len(txnFieldSpecs) {
		return txnFieldSpec{}, false
	}
	return txnFieldSpecs[f], true
}

// TxnFieldNames are arguments to the 'txn' family of opcodes.
var TxnFieldNames [invalidTxnField]string

var txnFieldSpecByName = make(tfNameSpecMap, len(TxnFieldNames))

// simple interface used by doc generator for fields versioning
type tfNameSpecMap map[string]txnFieldSpec

func (s tfNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

type txnFieldSpec struct {
	field      TxnField
	ftype      StackType
	array      bool   // Is this an array field?
	version    uint64 // When this field become available to txn/gtxn. 0=always
	itxVersion uint64 // When this field become available to itxn_field. 0=never
	effects    bool   // Is this a field on the "effects"? That is, something in ApplyData
	doc        string
}

func (fs txnFieldSpec) Field() byte {
	return byte(fs.field)
}
func (fs txnFieldSpec) Type() StackType {
	return fs.ftype
}
func (fs txnFieldSpec) OpVersion() uint64 {
	return 0
}
func (fs txnFieldSpec) Version() uint64 {
	return fs.version
}
func (fs txnFieldSpec) Note() string {
	note := fs.doc
	if fs.effects {
		note = addExtra(note, "Application mode only")
	}
	return note
}

var txnFieldSpecs = [...]txnFieldSpec{
	{Sender, StackAddress, false, 0, 5, false, "32 byte address"},
	{Fee, StackUint64, false, 0, 5, false, "microalgos"},
	{FirstValid, StackUint64, false, 0, 0, false, "round number"},
	{FirstValidTime, StackUint64, false, randomnessVersion, 0, false, "UNIX timestamp of block before txn.FirstValid. Fails if negative"},
	{LastValid, StackUint64, false, 0, 0, false, "round number"},
	{Note, StackBytes, false, 0, 6, false, "Any data up to 1024 bytes"},
	{Lease, StackBytes32, false, 0, 0, false, "32 byte lease value"},
	{Receiver, StackAddress, false, 0, 5, false, "32 byte address"},
	{Amount, StackUint64, false, 0, 5, false, "microalgos"},
	{CloseRemainderTo, StackAddress, false, 0, 5, false, "32 byte address"},
	{VotePK, StackBytes32, false, 0, 6, false, "32 byte address"},
	{SelectionPK, StackBytes32, false, 0, 6, false, "32 byte address"},
	{VoteFirst, StackUint64, false, 0, 6, false, "The first round that the participation key is valid."},
	{VoteLast, StackUint64, false, 0, 6, false, "The last round that the participation key is valid."},
	{VoteKeyDilution, StackUint64, false, 0, 6, false, "Dilution for the 2-level participation key"},
	{Type, StackBytes, false, 0, 5, false, "Transaction type as bytes"},
	{TypeEnum, StackUint64, false, 0, 5, false, "Transaction type as integer"},
	{XferAsset, StackUint64, false, 0, 5, false, "Asset ID"},
	{AssetAmount, StackUint64, false, 0, 5, false, "value in Asset's units"},
	{AssetSender, StackAddress, false, 0, 5, false,
		"32 byte address. Source of assets if Sender is the Asset's Clawback address."},
	{AssetReceiver, StackAddress, false, 0, 5, false, "32 byte address"},
	{AssetCloseTo, StackAddress, false, 0, 5, false, "32 byte address"},
	{GroupIndex, StackUint64, false, 0, 0, false,
		"Position of this transaction within an atomic transaction group. A stand-alone transaction is implicitly element 0 in a group of 1"},
	{TxID, StackBytes32, false, 0, 0, false, "The computed ID for this transaction. 32 bytes."},
	{ApplicationID, StackUint64, false, 2, 6, false, "ApplicationID from ApplicationCall transaction"},
	{OnCompletion, StackUint64, false, 2, 6, false, "ApplicationCall transaction on completion action"},
	{ApplicationArgs, StackBytes, true, 2, 6, false,
		"Arguments passed to the application in the ApplicationCall transaction"},
	{NumAppArgs, StackUint64, false, 2, 0, false, "Number of ApplicationArgs"},
	{Accounts, StackAddress, true, 2, 6, false, "Accounts listed in the ApplicationCall transaction"},
	{NumAccounts, StackUint64, false, 2, 0, false, "Number of Accounts"},
	{ApprovalProgram, StackBytes, false, 2, 6, false, "Approval program"},
	{ClearStateProgram, StackBytes, false, 2, 6, false, "Clear state program"},
	{RekeyTo, StackAddress, false, 2, 6, false, "32 byte Sender's new AuthAddr"},
	{ConfigAsset, StackUint64, false, 2, 5, false, "Asset ID in asset config transaction"},
	{ConfigAssetTotal, StackUint64, false, 2, 5, false, "Total number of units of this asset created"},
	{ConfigAssetDecimals, StackUint64, false, 2, 5, false,
		"Number of digits to display after the decimal place when displaying the asset"},
	{ConfigAssetDefaultFrozen, StackBoolean, false, 2, 5, false,
		"Whether the asset's slots are frozen by default or not, 0 or 1"},
	{ConfigAssetUnitName, StackBytes, false, 2, 5, false, "Unit name of the asset"},
	{ConfigAssetName, StackBytes, false, 2, 5, false, "The asset name"},
	{ConfigAssetURL, StackBytes, false, 2, 5, false, "URL"},
	{ConfigAssetMetadataHash, StackBytes32, false, 2, 5, false,
		"32 byte commitment to unspecified asset metadata"},
	{ConfigAssetManager, StackAddress, false, 2, 5, false, "32 byte address"},
	{ConfigAssetReserve, StackAddress, false, 2, 5, false, "32 byte address"},
	{ConfigAssetFreeze, StackAddress, false, 2, 5, false, "32 byte address"},
	{ConfigAssetClawback, StackAddress, false, 2, 5, false, "32 byte address"},
	{FreezeAsset, StackUint64, false, 2, 5, false, "Asset ID being frozen or un-frozen"},
	{FreezeAssetAccount, StackAddress, false, 2, 5, false,
		"32 byte address of the account whose asset slot is being frozen or un-frozen"},
	{FreezeAssetFrozen, StackBoolean, false, 2, 5, false, "The new frozen value, 0 or 1"},
	{Assets, StackUint64, true, 3, 6, false, "Foreign Assets listed in the ApplicationCall transaction"},
	{NumAssets, StackUint64, false, 3, 0, false, "Number of Assets"},
	{Applications, StackUint64, true, 3, 6, false, "Foreign Apps listed in the ApplicationCall transaction"},
	{NumApplications, StackUint64, false, 3, 0, false, "Number of Applications"},
	{GlobalNumUint, StackUint64, false, 3, 6, false, "Number of global state integers in ApplicationCall"},
	{GlobalNumByteSlice, StackUint64, false, 3, 6, false, "Number of global state byteslices in ApplicationCall"},
	{LocalNumUint, StackUint64, false, 3, 6, false, "Number of local state integers in ApplicationCall"},
	{LocalNumByteSlice, StackUint64, false, 3, 6, false, "Number of local state byteslices in ApplicationCall"},
	{ExtraProgramPages, StackUint64, false, 4, 6, false,
		"Number of additional pages for each of the application's approval and clear state programs. An ExtraProgramPages of 1 means 2048 more total bytes, or 1024 for each program."},
	{Nonparticipation, StackBoolean, false, 5, 6, false, "Marks an account nonparticipating for rewards"},

	// "Effects" Last two things are always going to: 0, true
	{Logs, StackBytes, true, 5, 0, true, "Log messages emitted by an application call (only with `itxn` in v5)"},
	{NumLogs, StackUint64, false, 5, 0, true, "Number of Logs (only with `itxn` in v5)"},
	{CreatedAssetID, StackUint64, false, 5, 0, true,
		"Asset ID allocated by the creation of an ASA (only with `itxn` in v5)"},
	{CreatedApplicationID, StackUint64, false, 5, 0, true,
		"ApplicationID allocated by the creation of an application (only with `itxn` in v5)"},
	{LastLog, StackBytes, false, 6, 0, true, "The last message emitted. Empty bytes if none were emitted"},

	// Not an effect. Just added after the effects fields.
	{StateProofPK, StackBytes, false, 6, 6, false, "64 byte state proof public key"},

	// Pseudo-fields to aid access to large programs (bigger than TEAL values)
	// reading in a txn seems not *super* useful, but setting in `itxn` is critical to inner app factories
	{ApprovalProgramPages, StackBytes, true, 7, 7, false, "Approval Program as an array of pages"},
	{NumApprovalProgramPages, StackUint64, false, 7, 0, false, "Number of Approval Program pages"},
	{ClearStateProgramPages, StackBytes, true, 7, 7, false, "ClearState Program as an array of pages"},
	{NumClearStateProgramPages, StackUint64, false, 7, 0, false, "Number of ClearState Program pages"},
}

// TxnFields contains info on the arguments to the txn* family of opcodes
var TxnFields = FieldGroup{
	"txn", "",
	TxnFieldNames[:],
	txnFieldSpecByName,
}

// TxnScalarFields narrows TxnFields to only have the names of scalar fetching opcodes
var TxnScalarFields = FieldGroup{
	"txn", "Fields (see [transaction reference](https://developer.algorand.org/docs/reference/transactions/))",
	txnScalarFieldNames(),
	txnFieldSpecByName,
}

// txnScalarFieldNames are txn field names that return scalars. Return value is
// a "sparse" slice, the names appear at their usual index, array slots are set
// to "".  They are laid out this way so that it is possible to get the name
// from the index value.
func txnScalarFieldNames() []string {
	names := make([]string, len(txnFieldSpecs))
	for i, fs := range txnFieldSpecs {
		if fs.array {
			names[i] = ""
		} else {
			names[i] = fs.field.String()
		}
	}
	return names
}

// TxnArrayFields narows TxnFields to only have the names of array fetching opcodes
var TxnArrayFields = FieldGroup{
	"txna", "Fields (see [transaction reference](https://developer.algorand.org/docs/reference/transactions/))",
	txnaFieldNames(),
	txnFieldSpecByName,
}

// txnaFieldNames are txn field names that return arrays. Return value is a
// "sparse" slice, the names appear at their usual index, non-array slots are
// set to "".  They are laid out this way so that it is possible to get the name
// from the index value.
func txnaFieldNames() []string {
	names := make([]string, len(txnFieldSpecs))
	for i, fs := range txnFieldSpecs {
		if fs.array {
			names[i] = fs.field.String()
		} else {
			names[i] = ""
		}
	}
	return names
}

// ItxnSettableFields collects info for itxn_field opcode
var ItxnSettableFields = FieldGroup{
	"itxn_field", "",
	itxnSettableFieldNames(),
	txnFieldSpecByName,
}

// itxnSettableFieldNames are txn field names that can be set by
// itxn_field. Return value is a "sparse" slice, the names appear at their usual
// index, unsettable slots are set to "".  They are laid out this way so that it is
// possible to get the name from the index value.
func itxnSettableFieldNames() []string {
	names := make([]string, len(txnFieldSpecs))
	for i, fs := range txnFieldSpecs {
		if fs.itxVersion == 0 {
			names[i] = ""
		} else {
			names[i] = fs.field.String()
		}
	}
	return names
}

var innerTxnTypes = map[string]uint64{
	string(protocol.PaymentTx):         5,
	string(protocol.KeyRegistrationTx): 6,
	string(protocol.AssetTransferTx):   5,
	string(protocol.AssetConfigTx):     5,
	string(protocol.AssetFreezeTx):     5,
	string(protocol.ApplicationCallTx): 6,
}

// TxnTypeNames is the values of Txn.Type in enum order
var TxnTypeNames = [...]string{
	string(protocol.UnknownTx),
	string(protocol.PaymentTx),
	string(protocol.KeyRegistrationTx),
	string(protocol.AssetConfigTx),
	string(protocol.AssetTransferTx),
	string(protocol.AssetFreezeTx),
	string(protocol.ApplicationCallTx),
}

// map txn type names (long and short) to index/enum value
var txnTypeMap = make(map[string]uint64)

// OnCompletionConstType is the same as transactions.OnCompletion
type OnCompletionConstType transactions.OnCompletion

const (
	// NoOp = transactions.NoOpOC
	NoOp = OnCompletionConstType(transactions.NoOpOC)
	// OptIn = transactions.OptInOC
	OptIn = OnCompletionConstType(transactions.OptInOC)
	// CloseOut = transactions.CloseOutOC
	CloseOut = OnCompletionConstType(transactions.CloseOutOC)
	// ClearState = transactions.ClearStateOC
	ClearState = OnCompletionConstType(transactions.ClearStateOC)
	// UpdateApplication = transactions.UpdateApplicationOC
	UpdateApplication = OnCompletionConstType(transactions.UpdateApplicationOC)
	// DeleteApplication = transactions.DeleteApplicationOC
	DeleteApplication = OnCompletionConstType(transactions.DeleteApplicationOC)
	// end of constants
	invalidOnCompletionConst = DeleteApplication + 1
)

// OnCompletionNames is the string names of Txn.OnCompletion, array index is the const value
var OnCompletionNames [invalidOnCompletionConst]string

// onCompletionMap maps symbolic name to uint64 for assembleInt
var onCompletionMap map[string]uint64

// GlobalField is an enum for `global` opcode
type GlobalField uint64

const (
	// MinTxnFee ConsensusParams.MinTxnFee
	MinTxnFee GlobalField = iota
	// MinBalance ConsensusParams.MinBalance
	MinBalance
	// MaxTxnLife ConsensusParams.MaxTxnLife
	MaxTxnLife
	// ZeroAddress [32]byte{0...}
	ZeroAddress
	// GroupSize len(txn group)
	GroupSize

	// v2

	// LogicSigVersion ConsensusParams.LogicSigVersion
	LogicSigVersion
	// Round basics.Round
	Round
	// LatestTimestamp uint64
	LatestTimestamp
	// CurrentApplicationID uint64
	CurrentApplicationID

	// v3

	// CreatorAddress [32]byte
	CreatorAddress

	// v5

	// CurrentApplicationAddress [32]byte
	CurrentApplicationAddress
	// GroupID [32]byte
	GroupID

	// v6

	// OpcodeBudget The remaining budget available for execution
	OpcodeBudget

	// CallerApplicationID The ID of the caller app, else 0
	CallerApplicationID

	// CallerApplicationAddress The Address of the caller app, else ZeroAddress
	CallerApplicationAddress

	// AssetCreateMinBalance is the additional minimum balance required to
	// create an asset (which also opts an account into that asset)
	AssetCreateMinBalance

	// AssetOptInMinBalance is the additional minimum balance required to opt in to an asset
	AssetOptInMinBalance

	// GenesisHash is the genesis hash for the network
	GenesisHash

	// PayoutsEnabled is whether block proposal payouts are enabled
	PayoutsEnabled

	// PayoutsGoOnlineFee is the fee required in a keyreg transaction to make an account incentive eligible
	PayoutsGoOnlineFee

	// PayoutsPercent is the percentage of transaction fees in a block that can be paid to the block proposer.
	PayoutsPercent

	// PayoutsMinBalance is the minimum algo balance an account must have to receive block payouts (in the agreement round).
	PayoutsMinBalance

	// PayoutsMaxBalance is the maximum algo balance an account can have to receive block payouts (in the agreement round).
	PayoutsMaxBalance

	invalidGlobalField // compile-time constant for number of fields
)

// GlobalFieldNames are arguments to the 'global' opcode
var GlobalFieldNames [invalidGlobalField]string

type globalFieldSpec struct {
	field   GlobalField
	ftype   StackType
	mode    RunMode
	version uint64
	doc     string
}

func (fs globalFieldSpec) Field() byte {
	return byte(fs.field)
}
func (fs globalFieldSpec) Type() StackType {
	return fs.ftype
}
func (fs globalFieldSpec) OpVersion() uint64 {
	return 0
}
func (fs globalFieldSpec) Version() uint64 {
	return fs.version
}
func (fs globalFieldSpec) Note() string {
	note := fs.doc
	if fs.mode == ModeApp {
		note = addExtra(note, "Application mode only.")
	}
	// There are no Signature mode only globals
	return note
}

const incentiveVersion = 11 // block fields, heartbeat

var globalFieldSpecs = [...]globalFieldSpec{
	// version 0 is the same as v1 (initial release)
	{MinTxnFee, StackUint64, ModeAny, 0, "microalgos"},
	{MinBalance, StackUint64, ModeAny, 0, "microalgos"},
	{MaxTxnLife, StackUint64, ModeAny, 0, "rounds"},
	{ZeroAddress, StackAddress, ModeAny, 0, "32 byte address of all zero bytes"},
	{GroupSize, StackUint64, ModeAny, 0,
		"Number of transactions in this atomic transaction group. At least 1"},
	{LogicSigVersion, StackUint64, ModeAny, 2, "Maximum supported version"},
	{Round, StackUint64, ModeApp, 2, "Current round number"},
	{LatestTimestamp, StackUint64, ModeApp, 2,
		"Last confirmed block UNIX timestamp. Fails if negative"},
	{CurrentApplicationID, StackUint64, ModeApp, 2, "ID of current application executing"},
	{CreatorAddress, StackAddress, ModeApp, 3,
		"Address of the creator of the current application"},
	{CurrentApplicationAddress, StackAddress, ModeApp, 5,
		"Address that the current application controls"},
	{GroupID, StackBytes32, ModeAny, 5,
		"ID of the transaction group. 32 zero bytes if the transaction is not part of a group."},
	{OpcodeBudget, StackUint64, ModeAny, 6,
		"The remaining cost that can be spent by opcodes in this program."},
	{CallerApplicationID, StackUint64, ModeApp, 6,
		"The application ID of the application that called this application. 0 if this application is at the top-level."},
	{CallerApplicationAddress, StackAddress, ModeApp, 6,
		"The application address of the application that called this application. ZeroAddress if this application is at the top-level."},
	{AssetCreateMinBalance, StackUint64, ModeAny, 10,
		"The additional minimum balance required to create (and opt-in to) an asset."},
	{AssetOptInMinBalance, StackUint64, ModeAny, 10,
		"The additional minimum balance required to opt-in to an asset."},
	{GenesisHash, StackBytes32, ModeAny, 10, "The Genesis Hash for the network."},

	{PayoutsEnabled, StackBoolean, ModeAny, incentiveVersion,
		"Whether block proposal payouts are enabled."},
	{PayoutsGoOnlineFee, StackUint64, ModeAny, incentiveVersion,
		"The fee required in a keyreg transaction to make an account incentive eligible."},
	{PayoutsPercent, StackUint64, ModeAny, incentiveVersion,
		"The percentage of transaction fees in a block that can be paid to the block proposer."},
	{PayoutsMinBalance, StackUint64, ModeAny, incentiveVersion,
		"The minimum balance an account must have in the agreement round to receive block payouts in the proposal round."},
	{PayoutsMaxBalance, StackUint64, ModeAny, incentiveVersion,
		"The maximum balance an account can have in the agreement round to receive block payouts in the proposal round."},
}

func globalFieldSpecByField(f GlobalField) (globalFieldSpec, bool) {
	if int(f) >= len(globalFieldSpecs) {
		return globalFieldSpec{}, false
	}
	return globalFieldSpecs[f], true
}

var globalFieldSpecByName = make(gfNameSpecMap, len(GlobalFieldNames))

type gfNameSpecMap map[string]globalFieldSpec

func (s gfNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// GlobalFields has info on the global opcode's immediate
var GlobalFields = FieldGroup{
	"global", "Fields",
	GlobalFieldNames[:],
	globalFieldSpecByName,
}

// EcdsaCurve is an enum for `ecdsa_` opcodes
type EcdsaCurve int

const (
	// Secp256k1 curve for bitcoin/ethereum
	Secp256k1 EcdsaCurve = iota
	// Secp256r1 curve
	Secp256r1
	invalidEcdsaCurve // compile-time constant for number of fields
)

var ecdsaCurveNames [invalidEcdsaCurve]string

type ecdsaCurveSpec struct {
	field   EcdsaCurve
	version uint64
	doc     string
}

func (fs ecdsaCurveSpec) Field() byte {
	return byte(fs.field)
}
func (fs ecdsaCurveSpec) Type() StackType {
	return StackNone // Will not show, since all are untyped
}
func (fs ecdsaCurveSpec) OpVersion() uint64 {
	return 5
}
func (fs ecdsaCurveSpec) Version() uint64 {
	return fs.version
}
func (fs ecdsaCurveSpec) Note() string {
	return fs.doc
}

var ecdsaCurveSpecs = [...]ecdsaCurveSpec{
	{Secp256k1, 5, "secp256k1 curve, used in Bitcoin"},
	{Secp256r1, fidoVersion, "secp256r1 curve, NIST standard"},
}

func ecdsaCurveSpecByField(c EcdsaCurve) (ecdsaCurveSpec, bool) {
	if int(c) >= len(ecdsaCurveSpecs) {
		return ecdsaCurveSpec{}, false
	}
	return ecdsaCurveSpecs[c], true
}

var ecdsaCurveSpecByName = make(ecdsaCurveNameSpecMap, len(ecdsaCurveNames))

type ecdsaCurveNameSpecMap map[string]ecdsaCurveSpec

func (s ecdsaCurveNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// EcdsaCurves collects details about the constants used to describe EcdsaCurves
var EcdsaCurves = FieldGroup{
	"ECDSA", "Curves",
	ecdsaCurveNames[:],
	ecdsaCurveSpecByName,
}

// EcGroup is an enum for `ec_` opcodes
type EcGroup int

const (
	// BN254g1 is the G1 group of BN254
	BN254g1 EcGroup = iota
	// BN254g2 is the G2 group of BN254
	BN254g2
	// BLS12_381g1 specifies the G1 group of BLS 12-381
	BLS12_381g1
	// BLS12_381g2 specifies the G2 group of BLS 12-381
	BLS12_381g2
	invalidEcGroup // compile-time constant for number of fields
)

var ecGroupNames [invalidEcGroup]string

type ecGroupSpec struct {
	field EcGroup
	doc   string
}

func (fs ecGroupSpec) Field() byte {
	return byte(fs.field)
}
func (fs ecGroupSpec) Type() StackType {
	return StackNone // Will not show, since all are untyped
}
func (fs ecGroupSpec) OpVersion() uint64 {
	return pairingVersion
}
func (fs ecGroupSpec) Version() uint64 {
	return pairingVersion
}
func (fs ecGroupSpec) Note() string {
	return fs.doc
}

var ecGroupSpecs = [...]ecGroupSpec{
	{BN254g1, "G1 of the BN254 curve. Points encoded as 32 byte X following by 32 byte Y"},
	{BN254g2, "G2 of the BN254 curve. Points encoded as 64 byte X following by 64 byte Y"},
	{BLS12_381g1, "G1 of the BLS 12-381 curve. Points encoded as 48 byte X following by 48 byte Y"},
	{BLS12_381g2, "G2 of the BLS 12-381 curve. Points encoded as 96 byte X following by 96 byte Y"},
}

func ecGroupSpecByField(c EcGroup) (ecGroupSpec, bool) {
	if int(c) >= len(ecGroupSpecs) {
		return ecGroupSpec{}, false
	}
	return ecGroupSpecs[c], true
}

var ecGroupSpecByName = make(ecGroupNameSpecMap, len(ecGroupNames))

type ecGroupNameSpecMap map[string]ecGroupSpec

func (s ecGroupNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// EcGroups collects details about the constants used to describe EcGroups
var EcGroups = FieldGroup{
	"EC", "Groups",
	ecGroupNames[:],
	ecGroupSpecByName,
}

// Base64Encoding is an enum for the `base64decode` opcode
type Base64Encoding int

const (
	// URLEncoding represents the base64url encoding defined in https://www.rfc-editor.org/rfc/rfc4648.html
	URLEncoding Base64Encoding = iota
	// StdEncoding represents the standard encoding of the RFC
	StdEncoding
	invalidBase64Encoding // compile-time constant for number of fields
)

var base64EncodingNames [invalidBase64Encoding]string

type base64EncodingSpec struct {
	field   Base64Encoding
	version uint64
}

var base64EncodingSpecs = [...]base64EncodingSpec{
	{URLEncoding, 6},
	{StdEncoding, 6},
}

func base64EncodingSpecByField(e Base64Encoding) (base64EncodingSpec, bool) {
	if int(e) >= len(base64EncodingSpecs) {
		return base64EncodingSpec{}, false
	}
	return base64EncodingSpecs[e], true
}

var base64EncodingSpecByName = make(base64EncodingSpecMap, len(base64EncodingNames))

type base64EncodingSpecMap map[string]base64EncodingSpec

func (fs base64EncodingSpec) Field() byte {
	return byte(fs.field)
}
func (fs base64EncodingSpec) Type() StackType {
	return StackAny // Will not show in docs, since all are untyped
}
func (fs base64EncodingSpec) OpVersion() uint64 {
	return 6
}
func (fs base64EncodingSpec) Version() uint64 {
	return fs.version
}
func (fs base64EncodingSpec) Note() string {
	note := "" // no doc list?
	return note
}

func (s base64EncodingSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// Base64Encodings describes the base64_encode immediate
var Base64Encodings = FieldGroup{
	"base64", "Encodings",
	base64EncodingNames[:],
	base64EncodingSpecByName,
}

// JSONRefType is an enum for the `json_ref` opcode
type JSONRefType int

const (
	// JSONString represents string json value
	JSONString JSONRefType = iota
	// JSONUint64 represents uint64 json value
	JSONUint64
	// JSONObject represents json object
	JSONObject
	invalidJSONRefType // compile-time constant for number of fields
)

var jsonRefTypeNames [invalidJSONRefType]string

type jsonRefSpec struct {
	field   JSONRefType
	ftype   StackType
	version uint64
}

var jsonRefSpecs = [...]jsonRefSpec{
	{JSONString, StackBytes, fidoVersion},
	{JSONUint64, StackUint64, fidoVersion},
	{JSONObject, StackBytes, fidoVersion},
}

func jsonRefSpecByField(r JSONRefType) (jsonRefSpec, bool) {
	if int(r) >= len(jsonRefSpecs) {
		return jsonRefSpec{}, false
	}
	return jsonRefSpecs[r], true
}

var jsonRefSpecByName = make(jsonRefSpecMap, len(jsonRefTypeNames))

type jsonRefSpecMap map[string]jsonRefSpec

func (fs jsonRefSpec) Field() byte {
	return byte(fs.field)
}
func (fs jsonRefSpec) Type() StackType {
	return fs.ftype
}
func (fs jsonRefSpec) OpVersion() uint64 {
	return fidoVersion
}
func (fs jsonRefSpec) Version() uint64 {
	return fs.version
}
func (fs jsonRefSpec) Note() string {
	note := "" // no doc list?
	return note
}

func (s jsonRefSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// JSONRefTypes describes the json_ref immediate
var JSONRefTypes = FieldGroup{
	"json_ref", "Types",
	jsonRefTypeNames[:],
	jsonRefSpecByName,
}

// VrfStandard is an enum for the `vrf_verify` opcode
type VrfStandard int

const (
	// VrfAlgorand is the built-in VRF of the Algorand chain
	VrfAlgorand        VrfStandard = iota
	invalidVrfStandard             // compile-time constant for number of fields
)

var vrfStandardNames [invalidVrfStandard]string

type vrfStandardSpec struct {
	field   VrfStandard
	version uint64
}

var vrfStandardSpecs = [...]vrfStandardSpec{
	{VrfAlgorand, randomnessVersion},
}

func vrfStandardSpecByField(r VrfStandard) (vrfStandardSpec, bool) {
	if int(r) >= len(vrfStandardSpecs) {
		return vrfStandardSpec{}, false
	}
	return vrfStandardSpecs[r], true
}

var vrfStandardSpecByName = make(vrfStandardSpecMap, len(vrfStandardNames))

type vrfStandardSpecMap map[string]vrfStandardSpec

func (s vrfStandardSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

func (fs vrfStandardSpec) Field() byte {
	return byte(fs.field)
}

func (fs vrfStandardSpec) Type() StackType {
	return StackNone // Will not show, since all are the same
}

func (fs vrfStandardSpec) OpVersion() uint64 {
	return randomnessVersion
}

func (fs vrfStandardSpec) Version() uint64 {
	return fs.version
}

func (fs vrfStandardSpec) Note() string {
	note := "" // no doc list?
	return note
}

func (s vrfStandardSpecMap) SpecByName(name string) FieldSpec {
	return s[name]
}

// VrfStandards describes the json_ref immediate
var VrfStandards = FieldGroup{
	"vrf_verify", "Standards",
	vrfStandardNames[:],
	vrfStandardSpecByName,
}

// BlockField is an enum for the `block` opcode
type BlockField int

const (
	// BlkSeed is the Block's vrf seed
	BlkSeed BlockField = iota
	// BlkTimestamp is the Block's timestamp, seconds from epoch
	BlkTimestamp
	invalidBlockField // compile-time constant for number of fields
)

var blockFieldNames [invalidBlockField]string

type blockFieldSpec struct {
	field   BlockField
	ftype   StackType
	version uint64
}

var blockFieldSpecs = [...]blockFieldSpec{
	{BlkSeed, StackBytes, randomnessVersion},
	{BlkTimestamp, StackUint64, randomnessVersion},
}

func blockFieldSpecByField(r BlockField) (blockFieldSpec, bool) {
	if int(r) >= len(blockFieldSpecs) {
		return blockFieldSpec{}, false
	}
	return blockFieldSpecs[r], true
}

var blockFieldSpecByName = make(blockFieldSpecMap, len(blockFieldNames))

type blockFieldSpecMap map[string]blockFieldSpec

func (s blockFieldSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

func (fs blockFieldSpec) Field() byte {
	return byte(fs.field)
}

func (fs blockFieldSpec) Type() StackType {
	return fs.ftype
}

func (fs blockFieldSpec) OpVersion() uint64 {
	return randomnessVersion
}

func (fs blockFieldSpec) Version() uint64 {
	return fs.version
}

func (fs blockFieldSpec) Note() string {
	return ""
}

func (s blockFieldSpecMap) SpecByName(name string) FieldSpec {
	return s[name]
}

// BlockFields describes the json_ref immediate
var BlockFields = FieldGroup{
	"block", "Fields",
	blockFieldNames[:],
	blockFieldSpecByName,
}

// AssetHoldingField is an enum for `asset_holding_get` opcode
type AssetHoldingField int

const (
	// AssetBalance AssetHolding.Amount
	AssetBalance AssetHoldingField = iota
	// AssetFrozen AssetHolding.Frozen
	AssetFrozen
	invalidAssetHoldingField // compile-time constant for number of fields
)

var assetHoldingFieldNames [invalidAssetHoldingField]string

type assetHoldingFieldSpec struct {
	field   AssetHoldingField
	ftype   StackType
	version uint64
	doc     string
}

func (fs assetHoldingFieldSpec) Field() byte {
	return byte(fs.field)
}
func (fs assetHoldingFieldSpec) Type() StackType {
	return fs.ftype
}
func (fs assetHoldingFieldSpec) OpVersion() uint64 {
	return 2
}
func (fs assetHoldingFieldSpec) Version() uint64 {
	return fs.version
}
func (fs assetHoldingFieldSpec) Note() string {
	return fs.doc
}

var assetHoldingFieldSpecs = [...]assetHoldingFieldSpec{
	{AssetBalance, StackUint64, 2, "Amount of the asset unit held by this account"},
	{AssetFrozen, StackBoolean, 2, "Is the asset frozen or not"},
}

func assetHoldingFieldSpecByField(f AssetHoldingField) (assetHoldingFieldSpec, bool) {
	if int(f) >= len(assetHoldingFieldSpecs) {
		return assetHoldingFieldSpec{}, false
	}
	return assetHoldingFieldSpecs[f], true
}

var assetHoldingFieldSpecByName = make(ahfNameSpecMap, len(assetHoldingFieldNames))

type ahfNameSpecMap map[string]assetHoldingFieldSpec

func (s ahfNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// AssetHoldingFields describes asset_holding_get's immediates
var AssetHoldingFields = FieldGroup{
	"asset_holding", "Fields",
	assetHoldingFieldNames[:],
	assetHoldingFieldSpecByName,
}

// AssetParamsField is an enum for `asset_params_get` opcode
type AssetParamsField int

const (
	// AssetTotal AssetParams.Total
	AssetTotal AssetParamsField = iota
	// AssetDecimals AssetParams.Decimals
	AssetDecimals
	// AssetDefaultFrozen AssetParams.AssetDefaultFrozen
	AssetDefaultFrozen
	// AssetUnitName AssetParams.UnitName
	AssetUnitName
	// AssetName AssetParams.AssetName
	AssetName
	// AssetURL AssetParams.URL
	AssetURL
	// AssetMetadataHash AssetParams.MetadataHash
	AssetMetadataHash
	// AssetManager AssetParams.Manager
	AssetManager
	// AssetReserve AssetParams.Reserve
	AssetReserve
	// AssetFreeze AssetParams.Freeze
	AssetFreeze
	// AssetClawback AssetParams.Clawback
	AssetClawback

	// AssetCreator is not *in* the Params, but it is uniquely determined.
	AssetCreator

	invalidAssetParamsField // compile-time constant for number of fields
)

var assetParamsFieldNames [invalidAssetParamsField]string

type assetParamsFieldSpec struct {
	field   AssetParamsField
	ftype   StackType
	version uint64
	doc     string
}

func (fs assetParamsFieldSpec) Field() byte {
	return byte(fs.field)
}
func (fs assetParamsFieldSpec) Type() StackType {
	return fs.ftype
}
func (fs assetParamsFieldSpec) OpVersion() uint64 {
	return 2
}
func (fs assetParamsFieldSpec) Version() uint64 {
	return fs.version
}
func (fs assetParamsFieldSpec) Note() string {
	return fs.doc
}

var assetParamsFieldSpecs = [...]assetParamsFieldSpec{
	{AssetTotal, StackUint64, 2, "Total number of units of this asset"},
	{AssetDecimals, StackUint64, 2, "See AssetParams.Decimals"},
	{AssetDefaultFrozen, StackBoolean, 2, "Frozen by default or not"},
	{AssetUnitName, StackBytes, 2, "Asset unit name"},
	{AssetName, StackBytes, 2, "Asset name"},
	{AssetURL, StackBytes, 2, "URL with additional info about the asset"},
	{AssetMetadataHash, StackBytes32, 2, "Arbitrary commitment"},
	{AssetManager, StackAddress, 2, "Manager address"},
	{AssetReserve, StackAddress, 2, "Reserve address"},
	{AssetFreeze, StackAddress, 2, "Freeze address"},
	{AssetClawback, StackAddress, 2, "Clawback address"},
	{AssetCreator, StackAddress, 5, "Creator address"},
}

func assetParamsFieldSpecByField(f AssetParamsField) (assetParamsFieldSpec, bool) {
	if int(f) >= len(assetParamsFieldSpecs) {
		return assetParamsFieldSpec{}, false
	}
	return assetParamsFieldSpecs[f], true
}

var assetParamsFieldSpecByName = make(apfNameSpecMap, len(assetParamsFieldNames))

type apfNameSpecMap map[string]assetParamsFieldSpec

func (s apfNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// AssetParamsFields describes asset_params_get's immediates
var AssetParamsFields = FieldGroup{
	"asset_params", "Fields",
	assetParamsFieldNames[:],
	assetParamsFieldSpecByName,
}

// AppParamsField is an enum for `app_params_get` opcode
type AppParamsField int

const (
	// AppApprovalProgram AppParams.ApprovalProgram
	AppApprovalProgram AppParamsField = iota
	// AppClearStateProgram AppParams.ClearStateProgram
	AppClearStateProgram
	// AppGlobalNumUint AppParams.StateSchemas.GlobalStateSchema.NumUint
	AppGlobalNumUint
	// AppGlobalNumByteSlice AppParams.StateSchemas.GlobalStateSchema.NumByteSlice
	AppGlobalNumByteSlice
	// AppLocalNumUint AppParams.StateSchemas.LocalStateSchema.NumUint
	AppLocalNumUint
	// AppLocalNumByteSlice AppParams.StateSchemas.LocalStateSchema.NumByteSlice
	AppLocalNumByteSlice
	// AppExtraProgramPages AppParams.ExtraProgramPages
	AppExtraProgramPages

	// AppCreator is not *in* the Params, but it is uniquely determined.
	AppCreator

	// AppAddress is also not *in* the Params, but can be derived
	AppAddress

	invalidAppParamsField // compile-time constant for number of fields
)

var appParamsFieldNames [invalidAppParamsField]string

type appParamsFieldSpec struct {
	field   AppParamsField
	ftype   StackType
	version uint64
	doc     string
}

func (fs appParamsFieldSpec) Field() byte {
	return byte(fs.field)
}
func (fs appParamsFieldSpec) Type() StackType {
	return fs.ftype
}
func (fs appParamsFieldSpec) OpVersion() uint64 {
	return 5
}
func (fs appParamsFieldSpec) Version() uint64 {
	return fs.version
}
func (fs appParamsFieldSpec) Note() string {
	return fs.doc
}

var appParamsFieldSpecs = [...]appParamsFieldSpec{
	{AppApprovalProgram, StackBytes, 5, "Bytecode of Approval Program"},
	{AppClearStateProgram, StackBytes, 5, "Bytecode of Clear State Program"},
	{AppGlobalNumUint, StackUint64, 5, "Number of uint64 values allowed in Global State"},
	{AppGlobalNumByteSlice, StackUint64, 5, "Number of byte array values allowed in Global State"},
	{AppLocalNumUint, StackUint64, 5, "Number of uint64 values allowed in Local State"},
	{AppLocalNumByteSlice, StackUint64, 5, "Number of byte array values allowed in Local State"},
	{AppExtraProgramPages, StackUint64, 5, "Number of Extra Program Pages of code space"},
	{AppCreator, StackAddress, 5, "Creator address"},
	{AppAddress, StackAddress, 5, "Address for which this application has authority"},
}

func appParamsFieldSpecByField(f AppParamsField) (appParamsFieldSpec, bool) {
	if int(f) >= len(appParamsFieldSpecs) {
		return appParamsFieldSpec{}, false
	}
	return appParamsFieldSpecs[f], true
}

var appParamsFieldSpecByName = make(appNameSpecMap, len(appParamsFieldNames))

// simple interface used by doc generator for fields versioning
type appNameSpecMap map[string]appParamsFieldSpec

func (s appNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// AppParamsFields describes app_params_get's immediates
var AppParamsFields = FieldGroup{
	"app_params", "Fields",
	appParamsFieldNames[:],
	appParamsFieldSpecByName,
}

// AcctParamsField is an enum for `acct_params_get` opcode
type AcctParamsField int

const (
	// AcctBalance is the balance, with pending rewards
	AcctBalance AcctParamsField = iota
	// AcctMinBalance is algos needed for this accounts apps and assets
	AcctMinBalance
	// AcctAuthAddr is the rekeyed address if any, else ZeroAddress
	AcctAuthAddr

	// AcctTotalNumUint is the count of all uints from created global apps or opted in locals
	AcctTotalNumUint
	// AcctTotalNumByteSlice is the count of all byte slices from created global apps or opted in locals
	AcctTotalNumByteSlice

	// AcctTotalExtraAppPages is the extra code pages across all apps
	AcctTotalExtraAppPages

	// AcctTotalAppsCreated is the number of apps created by this account
	AcctTotalAppsCreated
	// AcctTotalAppsOptedIn is the number of apps opted in by this account
	AcctTotalAppsOptedIn
	// AcctTotalAssetsCreated is the number of ASAs created by this account
	AcctTotalAssetsCreated
	// AcctTotalAssets is the number of ASAs opted in by this account (always includes AcctTotalAssetsCreated)
	AcctTotalAssets
	// AcctTotalBoxes is the number of boxes created by the app this account is associated with
	AcctTotalBoxes
	// AcctTotalBoxBytes is the number of bytes in all boxes of this app account
	AcctTotalBoxBytes

	// AcctTotalAppSchema - consider how to expose

	invalidAcctParamsField // compile-time constant for number of fields
)

var acctParamsFieldNames [invalidAcctParamsField]string

type acctParamsFieldSpec struct {
	field   AcctParamsField
	ftype   StackType
	version uint64
	doc     string
}

func (fs acctParamsFieldSpec) Field() byte {
	return byte(fs.field)
}
func (fs acctParamsFieldSpec) Type() StackType {
	return fs.ftype
}
func (fs acctParamsFieldSpec) OpVersion() uint64 {
	return 6
}
func (fs acctParamsFieldSpec) Version() uint64 {
	return fs.version
}
func (fs acctParamsFieldSpec) Note() string {
	return fs.doc
}

var acctParamsFieldSpecs = [...]acctParamsFieldSpec{
	{AcctBalance, StackUint64, 6, "Account balance in microalgos"},
	{AcctMinBalance, StackUint64, 6, "Minimum required balance for account, in microalgos"},
	{AcctAuthAddr, StackAddress, 6, "Address the account is rekeyed to."},

	{AcctTotalNumUint, StackUint64, 8, "The total number of uint64 values allocated by this account in Global and Local States."},
	{AcctTotalNumByteSlice, StackUint64, 8, "The total number of byte array values allocated by this account in Global and Local States."},
	{AcctTotalExtraAppPages, StackUint64, 8, "The number of extra app code pages used by this account."},
	{AcctTotalAppsCreated, StackUint64, 8, "The number of existing apps created by this account."},
	{AcctTotalAppsOptedIn, StackUint64, 8, "The number of apps this account is opted into."},
	{AcctTotalAssetsCreated, StackUint64, 8, "The number of existing ASAs created by this account."},
	{AcctTotalAssets, StackUint64, 8, "The numbers of ASAs held by this account (including ASAs this account created)."},
	{AcctTotalBoxes, StackUint64, boxVersion, "The number of existing boxes created by this account's app."},
	{AcctTotalBoxBytes, StackUint64, boxVersion, "The total number of bytes used by this account's app's box keys and values."},
}

func acctParamsFieldSpecByField(f AcctParamsField) (acctParamsFieldSpec, bool) {
	if int(f) >= len(acctParamsFieldSpecs) {
		return acctParamsFieldSpec{}, false
	}
	return acctParamsFieldSpecs[f], true
}

var acctParamsFieldSpecByName = make(acctNameSpecMap, len(acctParamsFieldNames))

type acctNameSpecMap map[string]acctParamsFieldSpec

func (s acctNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// AcctParamsFields describes acct_params_get's immediates
var AcctParamsFields = FieldGroup{
	"acct_params", "Fields",
	acctParamsFieldNames[:],
	acctParamsFieldSpecByName,
}

func init() {
	equal := func(x int, y int) {
		if x != y {
			panic(fmt.Sprintf("%d != %d", x, y))
		}
	}

	equal(len(txnFieldSpecs), len(TxnFieldNames))
	for i, s := range txnFieldSpecs {
		equal(int(s.field), i)
		TxnFieldNames[s.field] = s.field.String()
		txnFieldSpecByName[s.field.String()] = s
	}

	equal(len(globalFieldSpecs), len(GlobalFieldNames))
	for i, s := range globalFieldSpecs {
		equal(int(s.field), i)
		GlobalFieldNames[s.field] = s.field.String()
		globalFieldSpecByName[s.field.String()] = s
	}

	equal(len(ecdsaCurveSpecs), len(ecdsaCurveNames))
	for i, s := range ecdsaCurveSpecs {
		equal(int(s.field), i)
		ecdsaCurveNames[s.field] = s.field.String()
		ecdsaCurveSpecByName[s.field.String()] = s
	}

	equal(len(ecGroupSpecs), len(ecGroupNames))
	for i, s := range ecGroupSpecs {
		equal(int(s.field), i)
		ecGroupNames[s.field] = s.field.String()
		ecGroupSpecByName[s.field.String()] = s
	}

	equal(len(base64EncodingSpecs), len(base64EncodingNames))
	for i, s := range base64EncodingSpecs {
		equal(int(s.field), i)
		base64EncodingNames[i] = s.field.String()
		base64EncodingSpecByName[s.field.String()] = s
	}

	equal(len(jsonRefSpecs), len(jsonRefTypeNames))
	for i, s := range jsonRefSpecs {
		equal(int(s.field), i)
		jsonRefTypeNames[i] = s.field.String()
		jsonRefSpecByName[s.field.String()] = s
	}

	equal(len(vrfStandardSpecs), len(vrfStandardNames))
	for i, s := range vrfStandardSpecs {
		equal(int(s.field), i)
		vrfStandardNames[i] = s.field.String()
		vrfStandardSpecByName[s.field.String()] = s
	}

	equal(len(blockFieldSpecs), len(blockFieldNames))
	for i, s := range blockFieldSpecs {
		equal(int(s.field), i)
		blockFieldNames[i] = s.field.String()
		blockFieldSpecByName[s.field.String()] = s
	}

	equal(len(assetHoldingFieldSpecs), len(assetHoldingFieldNames))
	for i, s := range assetHoldingFieldSpecs {
		equal(int(s.field), i)
		assetHoldingFieldNames[i] = s.field.String()
		assetHoldingFieldSpecByName[s.field.String()] = s
	}

	equal(len(assetParamsFieldSpecs), len(assetParamsFieldNames))
	for i, s := range assetParamsFieldSpecs {
		equal(int(s.field), i)
		assetParamsFieldNames[i] = s.field.String()
		assetParamsFieldSpecByName[s.field.String()] = s
	}

	equal(len(appParamsFieldSpecs), len(appParamsFieldNames))
	for i, s := range appParamsFieldSpecs {
		equal(int(s.field), i)
		appParamsFieldNames[i] = s.field.String()
		appParamsFieldSpecByName[s.field.String()] = s
	}

	equal(len(acctParamsFieldSpecs), len(acctParamsFieldNames))
	for i, s := range acctParamsFieldSpecs {
		equal(int(s.field), i)
		acctParamsFieldNames[i] = s.field.String()
		acctParamsFieldSpecByName[s.field.String()] = s
	}

	txnTypeMap = make(map[string]uint64)
	for i, tt := range TxnTypeNames {
		txnTypeMap[tt] = uint64(i)
	}
	for k, v := range TypeNameDescriptions {
		txnTypeMap[v] = txnTypeMap[k]
	}

	onCompletionMap = make(map[string]uint64, len(OnCompletionNames))
	for oc := NoOp; oc < invalidOnCompletionConst; oc++ {
		symbol := oc.String()
		OnCompletionNames[oc] = symbol
		onCompletionMap[symbol] = uint64(oc)
	}

}

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Sender-0]
	_ = x[Fee-1]
	_ = x[FirstValid-2]
	_ = x[FirstValidTime-3]
	_ = x[LastValid-4]
	_ = x[Note-5]
	_ = x[Lease-6]
	_ = x[Receiver-7]
	_ = x[Amount-8]
	_ = x[CloseRemainderTo-9]
	_ = x[VotePK-10]
	_ = x[SelectionPK-11]
	_ = x[VoteFirst-12]
	_ = x[VoteLast-13]
	_ = x[VoteKeyDilution-14]
	_ = x[Type-15]
	_ = x[TypeEnum-16]
	_ = x[XferAsset-17]
	_ = x[AssetAmount-18]
	_ = x[AssetSender-19]
	_ = x[AssetReceiver-20]
	_ = x[AssetCloseTo-21]
	_ = x[GroupIndex-22]
	_ = x[TxID-23]
	_ = x[ApplicationID-24]
	_ = x[OnCompletion-25]
	_ = x[ApplicationArgs-26]
	_ = x[NumAppArgs-27]
	_ = x[Accounts-28]
	_ = x[NumAccounts-29]
	_ = x[ApprovalProgram-30]
	_ = x[ClearStateProgram-31]
	_ = x[RekeyTo-32]
	_ = x[ConfigAsset-33]
	_ = x[ConfigAssetTotal-34]
	_ = x[ConfigAssetDecimals-35]
	_ = x[ConfigAssetDefaultFrozen-36]
	_ = x[ConfigAssetUnitName-37]
	_ = x[ConfigAssetName-38]
	_ = x[ConfigAssetURL-39]
	_ = x[ConfigAssetMetadataHash-40]
	_ = x[ConfigAssetManager-41]
	_ = x[ConfigAssetReserve-42]
	_ = x[ConfigAssetFreeze-43]
	_ = x[ConfigAssetClawback-44]
	_ = x[FreezeAsset-45]
	_ = x[FreezeAssetAccount-46]
	_ = x[FreezeAssetFrozen-47]
	_ = x[Assets-48]
	_ = x[NumAssets-49]
	_ = x[Applications-50]
	_ = x[NumApplications-51]
	_ = x[GlobalNumUint-52]
	_ = x[GlobalNumByteSlice-53]
	_ = x[LocalNumUint-54]
	_ = x[LocalNumByteSlice-55]
	_ = x[ExtraProgramPages-56]
	_ = x[Nonparticipation-57]
	_ = x[Logs-58]
	_ = x[NumLogs-59]
	_ = x[CreatedAssetID-60]
	_ = x[CreatedApplicationID-61]
	_ = x[LastLog-62]
	_ = x[StateProofPK-63]
	_ = x[ApprovalProgramPages-64]
	_ = x[NumApprovalProgramPages-65]
	_ = x[ClearStateProgramPages-66]
	_ = x[NumClearStateProgramPages-67]
	_ = x[invalidTxnField-68]
}

const _TxnField_name = "SenderFeeFirstValidFirstValidTimeLastValidNoteLeaseReceiverAmountCloseRemainderToVotePKSelectionPKVoteFirstVoteLastVoteKeyDilutionTypeTypeEnumXferAssetAssetAmountAssetSenderAssetReceiverAssetCloseToGroupIndexTxIDApplicationIDOnCompletionApplicationArgsNumAppArgsAccountsNumAccountsApprovalProgramClearStateProgramRekeyToConfigAssetConfigAssetTotalConfigAssetDecimalsConfigAssetDefaultFrozenConfigAssetUnitNameConfigAssetNameConfigAssetURLConfigAssetMetadataHashConfigAssetManagerConfigAssetReserveConfigAssetFreezeConfigAssetClawbackFreezeAssetFreezeAssetAccountFreezeAssetFrozenAssetsNumAssetsApplicationsNumApplicationsGlobalNumUintGlobalNumByteSliceLocalNumUintLocalNumByteSliceExtraProgramPagesNonparticipationLogsNumLogsCreatedAssetIDCreatedApplicationIDLastLogStateProofPKApprovalProgramPagesNumApprovalProgramPagesClearStateProgramPagesNumClearStateProgramPagesinvalidTxnField"

var _TxnField_index = [...]uint16{0, 6, 9, 19, 33, 42, 46, 51, 59, 65, 81, 87, 98, 107, 115, 130, 134, 142, 151, 162, 173, 186, 198, 208, 212, 225, 237, 252, 262, 270, 281, 296, 313, 320, 331, 347, 366, 390, 409, 424, 438, 461, 479, 497, 514, 533, 544, 562, 579, 585, 594, 606, 621, 634, 652, 664, 681, 698, 714, 718, 725, 739, 759, 766, 778, 798, 821, 843, 868, 883}

func (i TxnField) String() string {
	if i < 0 || i >= TxnField(len(_TxnField_index)-1) {
		return "TxnField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TxnField_name[_TxnField_index[i]:_TxnField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[MinTxnFee-0]
	_ = x[MinBalance-1]
	_ = x[MaxTxnLife-2]
	_ = x[ZeroAddress-3]
	_ = x[GroupSize-4]
	_ = x[LogicSigVersion-5]
	_ = x[Round-6]
	_ = x[LatestTimestamp-7]
	_ = x[CurrentApplicationID-8]
	_ = x[CreatorAddress-9]
	_ = x[CurrentApplicationAddress-10]
	_ = x[GroupID-11]
	_ = x[OpcodeBudget-12]
	_ = x[CallerApplicationID-13]
	_ = x[CallerApplicationAddress-14]
	_ = x[AssetCreateMinBalance-15]
	_ = x[AssetOptInMinBalance-16]
	_ = x[GenesisHash-17]
	_ = x[PayoutsEnabled-18]
	_ = x[PayoutsGoOnlineFee-19]
	_ = x[PayoutsPercent-20]
	_ = x[PayoutsMinBalance-21]
	_ = x[PayoutsMaxBalance-22]
	_ = x[invalidGlobalField-23]
}

const _GlobalField_name = "MinTxnFeeMinBalanceMaxTxnLifeZeroAddressGroupSizeLogicSigVersionRoundLatestTimestampCurrentApplicationIDCreatorAddressCurrentApplicationAddressGroupIDOpcodeBudgetCallerApplicationIDCallerApplicationAddressAssetCreateMinBalanceAssetOptInMinBalanceGenesisHashPayoutsEnabledPayoutsGoOnlineFeePayoutsPercentPayoutsMinBalancePayoutsMaxBalanceinvalidGlobalField"

var _GlobalField_index = [...]uint16{0, 9, 19, 29, 40, 49, 64, 69, 84, 104, 118, 143, 150, 162, 181, 205, 226, 246, 257, 271, 289, 303, 320, 337, 355}

func (i GlobalField) String() string {
	if i >= GlobalField(len(_GlobalField_index)-1) {
		return "GlobalField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _GlobalField_name[_GlobalField_index[i]:_GlobalField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AssetTotal-0]
	_ = x[AssetDecimals-1]
	_ = x[AssetDefaultFrozen-2]
	_ = x[AssetUnitName-3]
	_ = x[AssetName-4]
	_ = x[AssetURL-5]
	_ = x[AssetMetadataHash-6]
	_ = x[AssetManager-7]
	_ = x[AssetReserve-8]
	_ = x[AssetFreeze-9]
	_ = x[AssetClawback-10]
	_ = x[AssetCreator-11]
	_ = x[invalidAssetParamsField-12]
}

const _AssetParamsField_name = "AssetTotalAssetDecimalsAssetDefaultFrozenAssetUnitNameAssetNameAssetURLAssetMetadataHashAssetManagerAssetReserveAssetFreezeAssetClawbackAssetCreatorinvalidAssetParamsField"

var _AssetParamsField_index = [...]uint8{0, 10, 23, 41, 54, 63, 71, 88, 100, 112, 123, 136, 148, 171}

func (i AssetParamsField) String() string {
	if i < 0 || i >= AssetParamsField(len(_AssetParamsField_index)-1) {
		return "AssetParamsField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AssetParamsField_name[_AssetParamsField_index[i]:_AssetParamsField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AppApprovalProgram-0]
	_ = x[AppClearStateProgram-1]
	_ = x[AppGlobalNumUint-2]
	_ = x[AppGlobalNumByteSlice-3]
	_ = x[AppLocalNumUint-4]
	_ = x[AppLocalNumByteSlice-5]
	_ = x[AppExtraProgramPages-6]
	_ = x[AppCreator-7]
	_ = x[AppAddress-8]
	_ = x[invalidAppParamsField-9]
}

const _AppParamsField_name = "AppApprovalProgramAppClearStateProgramAppGlobalNumUintAppGlobalNumByteSliceAppLocalNumUintAppLocalNumByteSliceAppExtraProgramPagesAppCreatorAppAddressinvalidAppParamsField"

var _AppParamsField_index = [...]uint8{0, 18, 38, 54, 75, 90, 110, 130, 140, 150, 171}

func (i AppParamsField) String() string {
	if i < 0 || i >= AppParamsField(len(_AppParamsField_index)-1) {
		return "AppParamsField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AppParamsField_name[_AppParamsField_index[i]:_AppParamsField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AcctBalance-0]
	_ = x[AcctMinBalance-1]
	_ = x[AcctAuthAddr-2]
	_ = x[AcctTotalNumUint-3]
	_ = x[AcctTotalNumByteSlice-4]
	_ = x[AcctTotalExtraAppPages-5]
	_ = x[AcctTotalAppsCreated-6]
	_ = x[AcctTotalAppsOptedIn-7]
	_ = x[AcctTotalAssetsCreated-8]
	_ = x[AcctTotalAssets-9]
	_ = x[AcctTotalBoxes-10]
	_ = x[AcctTotalBoxBytes-11]
	_ = x[invalidAcctParamsField-12]
}

const _AcctParamsField_name = "AcctBalanceAcctMinBalanceAcctAuthAddrAcctTotalNumUintAcctTotalNumByteSliceAcctTotalExtraAppPagesAcctTotalAppsCreatedAcctTotalAppsOptedInAcctTotalAssetsCreatedAcctTotalAssetsAcctTotalBoxesAcctTotalBoxBytesinvalidAcctParamsField"

var _AcctParamsField_index = [...]uint8{0, 11, 25, 37, 53, 74, 96, 116, 136, 158, 173, 187, 204, 226}

func (i AcctParamsField) String() string {
	if i < 0 || i >= AcctParamsField(len(_AcctParamsField_index)-1) {
		return "AcctParamsField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AcctParamsField_name[_AcctParamsField_index[i]:_AcctParamsField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AssetBalance-0]
	_ = x[AssetFrozen-1]
	_ = x[invalidAssetHoldingField-2]
}

const _AssetHoldingField_name = "AssetBalanceAssetFrozeninvalidAssetHoldingField"

var _AssetHoldingField_index = [...]uint8{0, 12, 23, 47}

func (i AssetHoldingField) String() string {
	if i < 0 || i >= AssetHoldingField(len(_AssetHoldingField_index)-1) {
		return "AssetHoldingField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AssetHoldingField_name[_AssetHoldingField_index[i]:_AssetHoldingField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[NoOp-0]
	_ = x[OptIn-1]
	_ = x[CloseOut-2]
	_ = x[ClearState-3]
	_ = x[UpdateApplication-4]
	_ = x[DeleteApplication-5]
}

const _OnCompletionConstType_name = "NoOpOptInCloseOutClearStateUpdateApplicationDeleteApplication"

var _OnCompletionConstType_index = [...]uint8{0, 4, 9, 17, 27, 44, 61}

func (i OnCompletionConstType) String() string {
	if i >= OnCompletionConstType(len(_OnCompletionConstType_index)-1) {
		return "OnCompletionConstType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _OnCompletionConstType_name[_OnCompletionConstType_index[i]:_OnCompletionConstType_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Secp256k1-0]
	_ = x[Secp256r1-1]
	_ = x[invalidEcdsaCurve-2]
}

const _EcdsaCurve_name = "Secp256k1Secp256r1invalidEcdsaCurve"

var _EcdsaCurve_index = [...]uint8{0, 9, 18, 35}

func (i EcdsaCurve) String() string {
	if i < 0 || i >= EcdsaCurve(len(_EcdsaCurve_index)-1) {
		return "EcdsaCurve(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _EcdsaCurve_name[_EcdsaCurve_index[i]:_EcdsaCurve_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[BN254g1-0]
	_ = x[BN254g2-1]
	_ = x[BLS12_381g1-2]
	_ = x[BLS12_381g2-3]
	_ = x[invalidEcGroup-4]
}

const _EcGroup_name = "BN254g1BN254g2BLS12_381g1BLS12_381g2invalidEcGroup"

var _EcGroup_index = [...]uint8{0, 7, 14, 25, 36, 50}

func (i EcGroup) String() string {
	if i < 0 || i >= EcGroup(len(_EcGroup_index)-1) {
		return "EcGroup(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _EcGroup_name[_EcGroup_index[i]:_EcGroup_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[URLEncoding-0]
	_ = x[StdEncoding-1]
	_ = x[invalidBase64Encoding-2]
}

const _Base64Encoding_name = "URLEncodingStdEncodinginvalidBase64Encoding"

var _Base64Encoding_index = [...]uint8{0, 11, 22, 43}

func (i Base64Encoding) String() string {
	if i < 0 || i >= Base64Encoding(len(_Base64Encoding_index)-1) {
		return "Base64Encoding(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Base64Encoding_name[_Base64Encoding_index[i]:_Base64Encoding_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[JSONString-0]
	_ = x[JSONUint64-1]
	_ = x[JSONObject-2]
	_ = x[invalidJSONRefType-3]
}

const _JSONRefType_name = "JSONStringJSONUint64JSONObjectinvalidJSONRefType"

var _JSONRefType_index = [...]uint8{0, 10, 20, 30, 48}

func (i JSONRefType) String() string {
	if i < 0 || i >= JSONRefType(len(_JSONRefType_index)-1) {
		return "JSONRefType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _JSONRefType_name[_JSONRefType_index[i]:_JSONRefType_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[VrfAlgorand-0]
	_ = x[invalidVrfStandard-1]
}

const _VrfStandard_name = "VrfAlgorandinvalidVrfStandard"

var _VrfStandard_index = [...]uint8{0, 11, 29}

func (i VrfStandard) String() string {
	if i < 0 || i >= VrfStandard(len(_VrfStandard_index)-1) {
		return "VrfStandard(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _VrfStandard_name[_VrfStandard_index[i]:_VrfStandard_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[BlkSeed-0]
	_ = x[BlkTimestamp-1]
	_ = x[invalidBlockField-2]
}

const _BlockField_name = "BlkSeedBlkTimestampinvalidBlockField"

var _BlockField_index = [...]uint8{0, 7, 19, 36}

func (i BlockField) String() string {
	if i < 0 || i >= BlockField(len(_BlockField_index)-1) {
		return "BlockField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _BlockField_name[_BlockField_index[i]:_BlockField_index[i+1]]
}

// TypeNameDescriptions contains extra description about a low level
// protocol transaction Type string, and provide a friendlier type
// constant name in assembler.
var TypeNameDescriptions = map[string]string{
	string(protocol.UnknownTx):         "Unknown type. Invalid transaction",
	string(protocol.PaymentTx):         "Payment",
	string(protocol.KeyRegistrationTx): "KeyRegistration",
	string(protocol.AssetConfigTx):     "AssetConfig",
	string(protocol.AssetTransferTx):   "AssetTransfer",
	string(protocol.AssetFreezeTx):     "AssetFreeze",
	string(protocol.ApplicationCallTx): "ApplicationCall",
}
