package parser

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
	"strings"
)

type indentString interface {
	String(indent int) string
}

func tabs(i int) string {
	return strings.Repeat("\t", i)
}

func (p *Program) String(indent int) string {
	return p.Block.String(indent) + "\n"
}

func (b *Block) String(indent int) string {
	stmts := make([]string, len(b.Statements))
	for i, stmt := range b.Statements {
		stmts[i] = stmt.String(indent)
	}
	if b.Return != nil {
		stmts = append(stmts, b.Return.String(indent))
	} else if b.Throw != nil {
		stmts = append(stmts, b.Throw.String(indent))
	}
	return strings.Join(stmts, "")
}

func (e *Elif) String(indent int) string {
	return fmt.Sprintf("%selif %s then\n%s", tabs(indent), e.Condition.String(0), e.Block.String(indent + 1))
}

func (i *IfElifElse) String(indent int) string {
	ifElifElse := fmt.Sprintf("%sif %s then\n%s", tabs(indent), i.IfCondition.String(0), i.IfBlock.String(indent + 1))
	if len(i.Elifs) > 0 {
		elifs := make([]string, len(i.Elifs))
		for e, elif := range i.Elifs {
			elifs[e] = elif.String(indent)
		}
		ifElifElse += strings.Join(elifs, "")
	}
	if i.Else != nil {
		ifElifElse += fmt.Sprintf("%selse\n%s", tabs(indent), i.Else.String(indent + 1))
	}
	ifElifElse += fmt.Sprintf("%send", tabs(indent))
	return ifElifElse
}

func (f *FunctionDefinition) String(indent int) string {
	return fmt.Sprintf("%sfunction %s%s", tabs(indent), f.JSONPath.String(0), f.Body.String(0))
}

func (tc *TryCatch) String(indent int) string {
	return fmt.Sprintf("%stry this\n%s%scatch as %s then\n%s%send", tabs(indent), tc.Try.String(indent + 1), tabs(indent), *tc.CatchAs, tc.Caught.String(indent + 1), tabs(indent))
}

func (b *Batch) String(indent int) string {
	return fmt.Sprintf("%sbatch this\n%s%send", tabs(indent), b.Block.String(indent + 1), tabs(indent))
}

func (f *ForEach) String(indent int) string {
	forEach := fmt.Sprintf("%sfor %s", tabs(indent), *f.Key)
	if f.Value != nil {
		forEach += fmt.Sprintf(", %s", *f.Value)
	}
	return forEach + fmt.Sprintf(" in %s do\n%s%send", f.In.String(0), f.Block.String(indent + 1), tabs(indent))
}

func (f *For) String(indent int) string {
	forLoop := fmt.Sprintf("%sfor %s; %s", tabs(indent), f.Var.String(0), f.Condition.String(0))
	if f.Step != nil {
		forLoop += fmt.Sprintf("; %s", f.Step.String(0))
	}
	return forLoop + fmt.Sprintf(" do\n%s%send", f.Block.String(indent + 1), tabs(indent))
}

func (w *While) String(indent int) string {
	return fmt.Sprintf("%swhile %s do\n%s\n%send", tabs(indent), w.Condition.String(0), w.Block.String(indent + 1), tabs(indent))
}

func (t *TestStatement) String(indent int) string {
	return fmt.Sprintf("%stest %s", tabs(indent), t.Expression.String(0))
}

func (a *Assignment) String(indent int) string {
	return fmt.Sprintf("%s%s = %s", tabs(indent), a.JSONPath.String(0), a.Value.String(0))
}

func (s *Statement) String(indent int) string {
	stmt := ""
	switch {
	case s.Assignment != nil:
		stmt = s.Assignment.String(indent)
	case s.FunctionCall != nil:
		stmt = s.FunctionCall.String(indent)
	case s.MethodCall != nil:
		stmt = s.MethodCall.String(indent)
	case s.Break != nil:
		stmt = "break"
	case s.Test != nil:
		stmt = s.Test.String(indent)
	case s.While != nil:
		stmt = s.While.String(indent)
	case s.For != nil:
		stmt = s.For.String(indent)
	case s.ForEach != nil:
		stmt = s.ForEach.String(indent)
	case s.Batch != nil:
		stmt = s.Batch.String(indent)
	case s.TryCatch != nil:
		stmt = s.TryCatch.String(indent)
	case s.FunctionDefinition != nil:
		stmt = s.FunctionDefinition.String(indent)
	case s.IfElifElse != nil:
		stmt = s.IfElifElse.String(indent)
	default:
		break
	}
	return stmt + ";\n"
}

func (p *Part) String(indent int) string {
	part := *p.Property
	if len(p.Indices) > 0 {
		expressions := make([]string, len(p.Indices))
		for i, expression := range p.Indices {
			expressions[i] = fmt.Sprintf("[%s]", expression.String(0))
		}
		part += strings.Join(expressions, "")
	}
	return part
}

func (j *JSONPath) String(indent int) string {
	parts := make([]string, len(j.Parts))
	for i, part := range j.Parts {
		parts[i] = part.String(0)
	}
	return strings.Join(parts, ".")
}

func (r *ReturnStatement) String(indent int) string {
	returnStmt := fmt.Sprintf("%sreturn", tabs(indent))
	if r.Value != nil {
		returnStmt += fmt.Sprintf(" %s", r.Value.String(0))
	}
	return returnStmt + ";\n"
}

func (t *ThrowStatement) String(indent int) string {
	throwStmt := fmt.Sprintf("%sthrow", tabs(indent))
	if t.Value != nil {
		throwStmt += fmt.Sprintf(" %s", t.Value.String(0))
	}
	return throwStmt + ";\n"
}

func (f *FunctionCall) String(indent int) string {
	functionCall := fmt.Sprintf("%s$%s(", tabs(indent), f.JSONPath.String(0))
	if len(f.Arguments) > 0 {
		args := make([]string, len(f.Arguments))
		for i, arg := range f.Arguments {
			args[i] = arg.String(0)
		}
		functionCall += strings.Join(args, ", ")
	}
	return functionCall + ")"
}

func (m *MethodCall) String(indent int) string {
	methodCall := fmt.Sprintf("%s$%s(", tabs(indent), methodNameMap[m.Method])
	if len(m.Arguments) > 0 {
		args := make([]string, len(m.Arguments))
		for i, arg := range m.Arguments {
			args[i] = arg.String(0)
		}
		methodCall += strings.Join(args, ", ")
	}
	return methodCall + ")"
}

func (fb *FunctionBody) String(indent int) string {
	params := make([]string, len(fb.Parameters))
	for i, param := range fb.Parameters {
		params[i] = param.String(0)
	}
	return fmt.Sprintf("(%s)\n%send", strings.Join(params, ", "), fb.Block.String(indent + 1))
}

func termString(indent int, factor indentString, next []indentString) string {
	expression := fmt.Sprintf("%s%s", tabs(indent), factor.String(0))
	if len(next) > 0 {
		nextStrs := make([]string, len(next))
		for i, n := range next {
			nextStrs[i] = n.String(0)
		}
		expression += " " + strings.Join(nextStrs, " ")
	}
	return expression
}

func precString(operator eval.Operator, factor indentString) string {
	return fmt.Sprintf("%s %s", operator.String(), factor.String(0))
}

func (e *Expression) String(indent int) string {
	next := make([]indentString, len(e.Right))
	for i, n := range e.Right {
		next[i] = n
	}
	return termString(indent, e.Left, next)
}

func (p5 *Prec5) String(indent int) string {
	return precString(p5.Operator, p5.Factor)
}

func (p5t *Prec5Term) String(indent int) string {
	next := make([]indentString, len(p5t.Right))
	for i, n := range p5t.Right {
		next[i] = n
	}
	return termString(indent, p5t.Left, next)
}

func (p4 *Prec4) String(indent int) string {
	return precString(p4.Operator, p4.Factor)
}

func (p4t *Prec4Term) String(indent int) string {
	next := make([]indentString, len(p4t.Right))
	for i, n := range p4t.Right {
		next[i] = n
	}
	return termString(indent, p4t.Left, next)
}

func (p3 *Prec3) String(indent int) string {
	return precString(p3.Operator, p3.Factor)
}

func (p3t *Prec3Term) String(indent int) string {
	next := make([]indentString, len(p3t.Right))
	for i, n := range p3t.Right {
		next[i] = n
	}
	return termString(indent, p3t.Left, next)
}

func (p2 *Prec2) String(indent int) string {
	return precString(p2.Operator, p2.Factor)
}

func (p2t *Prec2Term) String(indent int) string {
	next := make([]indentString, len(p2t.Right))
	for i, n := range p2t.Right {
		next[i] = n
	}
	return termString(indent, p2t.Left, next)
}

func (p1 *Prec1) String(indent int) string {
	return precString(p1.Operator, p1.Factor)
}

func (p1t *Prec1Term) String(indent int) string {
	next := make([]indentString, len(p1t.Right))
	for i, n := range p1t.Right {
		next[i] = n
	}
	return termString(indent, p1t.Left, next)
}

func (p0 *Prec0) String(indent int) string {
	return precString(p0.Operator, p0.Factor)
}

func (f *Factor) String(indent int) string {
	fac := ""
	switch {
	case f.Null != nil:
		fac = "null"
	case f.Boolean != nil:
		fac = fmt.Sprintf("%v", bool(*f.Boolean))
	case f.Number != nil:
		fac = fmt.Sprintf("%v", *f.Number)
	case f.StringLit != nil:
		fac = fmt.Sprintf("\"%s\"", *f.StringLit)
	case f.JSONPath != nil:
		fac = f.JSONPath.String(0)
	case f.JSON != nil:
		fac = f.JSON.String(0)
	case f.FunctionCall != nil:
		fac = f.FunctionCall.String(0)
	case f.MethodCall != nil:
		fac = f.MethodCall.String(0)
	case f.SubExpression != nil:
		fac = fmt.Sprintf("(%s)", f.SubExpression.String(0))
	default:
		panic("factor does not have any non-nil fields")
	}
	return fac
}

func (j *JSON) String(indent int) string {
	if j.Object != nil {
		return j.Object.String(indent)
	}
	return j.Array.String(indent)
}

func (o *Object) String(indent int) string {
	pairs := make([]string, len(o.Pairs))
	for i := range o.Pairs {
		pairs[i] = o.Pairs[i].String(0)
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

func (p *Pair) String(indent int) string {
	return fmt.Sprintf("%s: %s", p.Key.String(0), p.Value.String(0))
}

func (a *Array) String(indent int) string {
	elems := make([]string, len(a.Elements))
	for i := range a.Elements {
		elems[i] = a.Elements[i].String(0)
	}
	return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
}
