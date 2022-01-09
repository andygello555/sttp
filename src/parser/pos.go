package parser

import "github.com/alecthomas/participle/v2/lexer"

type positionable interface {
	GetPos() lexer.Position
}

func (p *Program) GetPos() lexer.Position { return p.Pos }
func (b *Block) GetPos() lexer.Position { return b.Pos }
func (i *IfElifElse) GetPos() lexer.Position { return i.Pos }
func (f *FunctionDefinition) GetPos() lexer.Position { return f.Pos }
func (tc *TryCatch) GetPos() lexer.Position { return tc.Pos }
func (b *Batch) GetPos() lexer.Position { return b.Pos }
func (f *ForEach) GetPos() lexer.Position { return f.Pos }
func (f *For) GetPos() lexer.Position { return f.Pos }
func (w *While) GetPos() lexer.Position { return w.Pos }
func (t *TestStatement) GetPos() lexer.Position { return t.Pos }
func (a *Assignment) GetPos() lexer.Position { return a.Pos }
func (s *Statement) GetPos() lexer.Position { return s.Pos }
func (j *JSONPath) GetPos() lexer.Position { return j.Pos }
func (r *ReturnStatement) GetPos() lexer.Position { return r.Pos }
func (t *ThrowStatement) GetPos() lexer.Position { return t.Pos }
func (f *FunctionCall) GetPos() lexer.Position { return f.Pos }
func (m *MethodCall) GetPos() lexer.Position { return m.Pos }
func (e *Expression) GetPos() lexer.Position { return e.Pos }
func (p5 *Prec5) GetPos() lexer.Position { return p5.Pos }
func (p5t *Prec5Term) GetPos() lexer.Position { return p5t.Pos }
func (p4 *Prec4) GetPos() lexer.Position { return p4.Pos }
func (p4t *Prec4Term) GetPos() lexer.Position { return p4t.Pos }
func (p3 *Prec3) GetPos() lexer.Position { return p3.Pos }
func (p3t *Prec3Term) GetPos() lexer.Position { return p3t.Pos }
func (p2 *Prec2) GetPos() lexer.Position { return p2.Pos }
func (p2t *Prec2Term) GetPos() lexer.Position { return p2t.Pos }
func (p1 *Prec1) GetPos() lexer.Position { return p1.Pos }
func (p1t *Prec1Term) GetPos() lexer.Position { return p1t.Pos }
func (p0 *Prec0) GetPos() lexer.Position { return p0.Pos }
func (f *Factor) GetPos() lexer.Position { return f.Pos }
func (j *JSON) GetPos() lexer.Position { return j.Pos }
func (n *Null) GetPos() lexer.Position { return lexer.Position{} }
func (b *Boolean) GetPos() lexer.Position { return lexer.Position{} }
