package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/dragmz/teal"

	"github.com/samber/lo"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

type args struct {
	path string
}

type compiler struct {
	labels map[string]*teal.LabelExpr
	store  map[any]uint8
	funcs  map[*ssa.Function]*teal.FuncExpr
	blocks map[*ssa.BasicBlock]*teal.NestedExpr
}

func (c *compiler) translateLookup(f *ssa.Function, l *ssa.Lookup) []teal.Expr {
	var body []teal.Expr

	ref, ok := c.store[l]
	if ok {
		body = append(body, teal.Load(ref))
	} else {
		switch v := l.X.(type) {
		case *ssa.UnOp:
			switch x := v.X.(type) {
			case *ssa.Global:
				switch x.Name() {
				case "Global":
					switch i := l.Index.(type) {
					case *ssa.Const:
						f := teal.GlobalField(i.Int64())
						body = append(body, teal.Global(f), c.addRef(l))
					default:
						panic("unsupported index type")
					}
				default:
					panic("unsupported global")
				}
			default:
				panic("unsupported lookup value type")
			}
		default:
			panic("unsupported lookup type")
		}
	}

	return body
}

func (c *compiler) translateToken(t token.Token) teal.Expr {
	switch t {
	case token.LSS:
		return teal.Lt
	case token.SUB:
		return teal.MinusOp
	case token.EQL:
		return teal.Eq
	case token.ADD:
		return teal.PlusOp
	case token.GTR:
		return teal.Gt
	case token.MUL:
		return teal.Mul
	default:
		panic(fmt.Sprintf("unsupported token: %#v", t.String()))
	}
}

func (c *compiler) getLabel(name string) *teal.LabelExpr {
	e, ok := c.labels[name]
	if !ok {
		e = teal.Label(name)
		c.labels[name] = e
	}

	return e
}

func (c *compiler) addRef(v ssa.Value) []teal.Expr {
	var body []teal.Expr

	refs := v.Referrers()
	if refs != nil && len(*refs) > 0 {
		index := uint8(len(c.store)) // TODO: take free index
		body = append(body, teal.Store(index))
		c.store[v] = index
	}

	return body
}

func (c *compiler) translateUnOp(m *ssa.Function, o *ssa.UnOp) []teal.Expr {
	var body []teal.Expr

	ref, ok := c.store[o]
	if ok {
		body = append(body, teal.Load(ref))
	} else {
		switch o.X.(type) {
		case *ssa.Global:
			return []teal.Expr{}
		}
		body = append(body, c.translateValue(m, o.X))
		body = append(body, c.translateToken(o.Op))
		body = append(body, c.addRef(m)...)
	}

	return body
}

func (c *compiler) translateBinOp(m *ssa.Function, o *ssa.BinOp) []teal.Expr {
	var body []teal.Expr

	ref, ok := c.store[o]
	if ok {
		body = append(body, teal.Load(ref))
	} else {
		body = append(body, c.translateValue(m, o.X))
		body = append(body, c.translateValue(m, o.Y))
		body = append(body, c.translateToken(o.Op))
		body = append(body, c.addRef(o)...)
	}

	return body
}

func (c *compiler) translateInstruction(f *ssa.Function, i ssa.Instruction) []teal.Expr {
	switch i := i.(type) {
	case *ssa.BinOp:
		return c.translateBinOp(f, i)
	case *ssa.UnOp:
		return c.translateUnOp(f, i)
	case *ssa.If:
		var body []teal.Expr

		switch cnd := i.Cond.(type) {
		case *ssa.Const:
			switch cnd.Value.Kind() {
			case constant.Bool:
				var v uint64
				if constant.BoolVal(cnd.Value) {
					v = 1
				} else {
					v = 0
				}
				body = []teal.Expr{teal.Int(v)}
			default:
				panic(fmt.Sprintf("unsupported value kind: %#v", cnd))
			}
		case *ssa.UnOp:
			body = append(body, c.translateUnOp(f, cnd)...)
		case *ssa.BinOp:
			body = append(body, c.translateBinOp(f, cnd)...)
		default:
			panic(fmt.Sprintf("unsupported condition: %#v", cnd))
		}

		succs := i.Block().Succs

		succ1, _ := c.translateBlock(f, succs[0])
		succ2, _ := c.translateBlock(f, succs[1])

		body = append(body, teal.Bnz(succ1.Label))
		body = append(body, teal.B(succ2.Label))

		return body
	case *ssa.Return:
		var body []teal.Expr
		for _, v := range i.Results {
			body = append(body, c.translateValue(f, v))
		}
		return body
	case *ssa.Call:
		return c.translateCall(f, i)
	case *ssa.Jump:
		block, _ := c.translateBlock(f, i.Block().Succs[0])
		return []teal.Expr{block}
	case *ssa.Lookup:
		return c.translateLookup(f, i)
	default:
		panic(fmt.Sprintf("unknown instr: %#v", i))
	}
}

func (c *compiler) translateCall(f *ssa.Function, i *ssa.Call) []teal.Expr {
	ref, ok := c.store[i]
	if ok {
		return []teal.Expr{teal.Load(ref)}
	} else {

		var body []teal.Expr
		for _, arg := range i.Call.Args {
			body = append(body, c.translateValue(f, arg))
		}

		switch f2 := i.Call.Value.(type) {
		case *ssa.Function:
			var funbody []teal.Expr
			m, ok := c.funcs[f2]
			if !ok {
				funbody = c.translateFunction(f2)
				m = c.funcs[f2]
			}

			body = append(body, teal.CallSub(m.Label))
			body = append(body, c.addRef(i)...)
			body = append(body, funbody...)
		case *ssa.Lookup:
			body = append(body, c.translateLookup(f, f2))
		default:
			panic("unsupported call type")
		}

		return body
	}
}

func (c *compiler) getBlockLabel(b *ssa.BasicBlock) *teal.LabelExpr {
	return c.getLabel(fmt.Sprintf("%s_blk_%d", b.Parent().Name(), b.Index))
}

func (c *compiler) translateFunction(m *ssa.Function) []teal.Expr {
	var body []teal.Expr

	end := c.getLabel(fmt.Sprintf("func_%s_end", m.Name()))

	var blocks []teal.Expr

	var proto *teal.ProtoExpr

	args := uint8(m.Signature.Params().Len())
	rets := uint8(m.Signature.Results().Len())

	if args != 0 || rets != 0 {
		proto = teal.Proto(args, rets)
	}
	f := &teal.FuncExpr{
		Label: c.getLabel(fmt.Sprintf("func_%s", m.Name())),
		Proto: proto,
	}

	c.funcs[m] = f

	for _, b := range m.Blocks {
		block, _ := c.translateBlock(m, b)
		blocks = append(blocks, block, teal.B(end))
	}

	blocks = append(blocks, end)

	f.Block = &teal.NestedExpr{
		Body: blocks,
	}

	body = append(body, f)

	return body
}

func (c *compiler) translateBlock(f *ssa.Function, b *ssa.BasicBlock) (*teal.NestedExpr, bool) {
	blk, ok := c.blocks[b]
	if ok {
		return blk, ok
	}

	var body []teal.Expr

	label := c.getBlockLabel(b)
	block := &teal.NestedExpr{
		Label: label,
	}

	c.blocks[b] = block

	for _, i := range b.Instrs {
		exprs := c.translateInstruction(f, i)
		body = append(body, exprs...)
	}

	block.Body = body

	return block, false
}

func (c *compiler) translateValue(f *ssa.Function, v ssa.Value) teal.Expr {
	switch v := v.(type) {
	case *ssa.Const:
		switch v.Value.Kind() {
		case constant.Int:
			return teal.Int(v.Uint64())
		case constant.Bool:
			if constant.BoolVal(v.Value) {
				return teal.Int(1)
			} else {
				return teal.Int(0)
			}
		default:
			panic(fmt.Sprintf("unexpected ssa type: %#v", v.Value.Kind()))
		}
	case *ssa.Parameter:
		_, n, ok := lo.FindIndexOf(f.Params, func(p *ssa.Parameter) bool {
			return p.Name() == v.Name()
		})

		if !ok {
			panic(fmt.Sprintf("parameter not found: %#v", v))
		}

		return teal.FrameDig(int8(-len(f.Params) + n))
	case *ssa.BinOp:
		return &teal.NestedExpr{Body: c.translateBinOp(f, v)}
	case *ssa.Call:
		return &teal.NestedExpr{Body: c.translateCall(f, v)}
	case *ssa.Global:
		panic("global")
	case *ssa.Lookup:
		return teal.Block(c.translateLookup(f, v))
	default:
		panic(fmt.Sprintf("unexpected ssa value: %#v", v))
	}
}

func (c *compiler) run(a args) error {

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, a.path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	files := []*ast.File{f}

	// Create the type-checker's package.
	pkg := types.NewPackage("app", "")

	// Type-check the package, load dependencies.
	// Create and build the SSA program.
	app, _, err := ssautil.BuildPackage(
		&types.Config{Importer: importer.Default()}, fset, pkg, files, ssa.SanityCheckFunctions)
	if err != nil {
		return err
	}

	var body []teal.Expr

	for _, m := range app.Members {
		switch m := m.(type) {
		case *ssa.Global:
			if m.Name() == "init$guard" {
				continue
			}
			switch t := m.Type().(type) {
			default:
			case *types.Pointer:
				panic(fmt.Sprintf("unknown type: %s", t.String()))
			}
		case *ssa.Function:
			if m.Name() == "init" {
				// todo: not supported yet
				continue
			}
			body = append(body, c.translateFunction(m)...)
		case *ssa.Type:
			fmt.Println("type", m.Name())
		case *ssa.NamedConst:
		default:
			fmt.Println("unknown", m.Name())
		}
	}

	source := teal.Compile(body)
	fmt.Println(source)

	return nil
}

func main() {
	var a args

	flag.StringVar(&a.path, "path", "", "")
	flag.Parse()

	c := &compiler{
		blocks: map[*ssa.BasicBlock]*teal.NestedExpr{},
		labels: map[string]*teal.LabelExpr{},
		store:  map[any]uint8{},
		funcs:  map[*ssa.Function]*teal.FuncExpr{},
	}

	err := c.run(a)
	if err != nil {
		panic(err)
	}
}
