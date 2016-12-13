// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Logical Terms

package terms

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type CList []*Compound

type Type int

const (
	AtomType Type = iota
	BoolType
	IntType
	FloatType
	StringType
	CompoundType
	ListType
	VariableType
)

type Vars []Variable

type Term interface {
	OccurVars() Vars
	String() string
	Type() Type
}

type Atom string
type Bool bool
type Int int
type Float float64
type String string

// type EnvMap map[int][]Bindings

type Compound struct {
	Functor string
	Id      *big.Int
	Prio    int
	//	EMap              *EnvMap
	occurVars         Vars
	identifyOccurVars bool
	IsDeleted         bool
	Args              []Term
	HasArgs           bool
}

type List []Term

type Variable struct {
	Name  string
	index *big.Int
}

func (t CList) OccurVars() Vars {
	occur := Vars{}
	for _, t2 := range t {
		occur = append(occur, t2.OccurVars()...)
	}
	return occur
}

func (t CList) String() string {
	args := []string{}
	for _, arg := range t {
		args = append(args, arg.String())
	}
	return "[" + strings.Join(args, ", ") + "]"
}

func (t CList) Type() Type {
	return ListType
}

func NewVariable(name string) Variable {
	return Variable{Name: name, index: big.NewInt(0)}
}

func IsNewVariable(v Variable) bool {
	return v.index.Cmp(big.NewInt(0)) == 0
}

func EqVars(v1, v2 Variable) bool {
	if v1.Name == v2.Name && v1.index.Cmp(v2.index) == 0 {
		return true
	}
	return false
}

func CopyCompound(c Compound) (c1 Compound) {
	c1 = Compound{
		Functor: c.Functor,
		Id:      c.Id,
		Prio:    c.Prio,
		//		EMap:              c.EMap,
		occurVars:         c.occurVars,
		identifyOccurVars: c.identifyOccurVars,
		IsDeleted:         c.IsDeleted,
		Args:              []Term{},
		HasArgs:           c.HasArgs,
	}
	args := []Term{}
	for _, a := range c.Args {
		args = append(args, a)
	}
	c1.Args = args
	return c1
}

type Bindings *BindEle
type BindEle struct {
	Var  Variable
	T    Term
	Next Bindings
}

func AddBinding(v Variable, t Term, b Bindings) Bindings {
	// fmt.Printf(" Add Binding %s-%d == %s \n", v.String(), v.index, t.String())
	return &BindEle{Var: v, T: t, Next: b}
}

func GetBinding(v Variable, b Bindings) (t Term, ok bool) {
	// fmt.Printf(" GetBinding %s-%d %v \n", v.String(), v.index, b)
	name := v.Name
	id := v.index
	if id == nil {
		for b != nil {
			if b.Var.Name == name && b.Var.index == nil {
				// fmt.Printf(" Binding found %s \n", b.T.String())
				return b.T, true
			}
			b = b.Next
			// fmt.Printf(" NextBinding %s %v \n", name, b)
		}
	} else {

		for b != nil {
			if b.Var.Name == name && b.Var.index != nil && b.Var.index.Cmp(id) == 0 {
				// fmt.Printf(" Binding found %s \n", b.T.String())
				return b.T, true
			}
			b = b.Next
			// fmt.Printf(" NextBinding %s %v \n", name, b)
		}
	}
	// fmt.Printf(" Binding not found \n")
	return nil, false
}

func isIn(v Variable, vl []Variable) bool {
	name := v.Name
	id := v.index
	if id != nil {
		for _, v2 := range vl {
			if v2.Name == name && v2.index != nil && v2.index.Cmp(id) == 0 {
				return true
			}
		}
		return false
	} else {
		for _, v2 := range vl {
			if v2.Name == name && v2.index == nil {
				return true
			}
		}

	}
	return false
}

func GetImplicitEquals(b Bindings) (cl List, ok bool) {
	cl = List{}
	ok = false
	vl := []Variable{}

	for b != nil {
		if b.T != nil && b.T.Type() == VariableType && !isIn(b.T.(Variable), vl) {
			v := b.T.(Variable)
			vl = append(vl, v)
			name := v.Name
			id := v.index
			vl2 := []Variable{b.Var}
			b2 := b.Next
			for b2 != nil {
				if b.T != nil && b.T.Type() == VariableType {
					v2 := b.T.(Variable)
					if v2.Name == name && v2.index.Cmp(id) == 0 && !isIn(b.Var, vl2) {
						vl2 = append(vl2, b.Var)
					}
				}
				b2 = b2.Next
			}
			l := len(vl2)
			for i := 1; i < l; i++ {
				cl = append(cl, Compound{Functor: "==", Args: []Term{vl2[0], vl2[i]}}) // Prio: 3

				ok = true
			}

		}
		b = b.Next
	}
	return cl, ok
}

/* old GetReserveBinding
	// fmt.Printf(" GetBinding %s-%d %v \n", v.String(), v.index, b)
	name := v.Name
	id := v.index
	if id == nil {
		for b != nil {
			if b.T != nil && b.T.Type() == VariableType && b.T.(Variable).Name == name && b.T.(Variable).index == nil {
				// fmt.Printf(" Binding found %s \n", b.T.String())
				return &b.Var, true
			}
			b = b.Next
			// fmt.Printf(" NextBinding %s %v \n", name, b)
		}
	} else {

		for b != nil {
			// fmt.Printf("---Reverse: b.T: %s, b.Var: %s var: %s \n", b.T, b.Var, name)
			if b.T != nil && b.T.Type() == VariableType && b.T.(Variable).Name == name && b.T.(Variable).index != nil && b.T.(Variable).index.Cmp(id) == 0 {
				// fmt.Printf(" Binding found %s \n", b.T.String())
				return &b.Var, true
			}
			b = b.Next
			// fmt.Printf(" NextBinding %s %v \n", name, b)
		}
	}
	// fmt.Printf(" Binding not found \n")
	return nil, false
}
*/

func (t Atom) Type() Type {
	return AtomType
}

func (t Bool) Type() Type {
	return BoolType
}

func (t Int) Type() Type {
	return IntType
}

func (t Float) Type() Type {
	return FloatType
}

func (t String) Type() Type {
	return StringType
}

func (t Compound) Type() Type {
	return CompoundType
}

func (t List) Type() Type {
	return ListType
}

func (t Variable) Type() Type {
	return VariableType
}

func (t Atom) String() string {
	return string(t)
}

func (t Bool) String() string {
	if t {
		return "true"
	} else {
		return "false"
	}
}

func (t Int) String() string {
	return strconv.Itoa(int(t))
}

func (t Float) String() string {
	return fmt.Sprintf("%f", t)
}

func (t String) String() string {
	return string(t)
}

func (t Compound) String() string {
	if t.Prio != 0 {
		prio := t.Prio
		f := t.Functor
		switch f {
		case "||", "&&", "in", "or", "div", "mod":
			f = " " + f + " "
		}
		switch t.Arity() {
		case 1:
			if t.Args[0].Type() == CompoundType {
				prio1 := t.Args[0].(Compound).Prio
				if prio1 == 0 {
					return f + t.Args[0].String()
				}
				if prio1 < prio {
					return f + "(" + t.Args[0].String() + ")"
				}
			}
			return f + t.Args[0].String()
		case 2:
			if t.Args[0].Type() == CompoundType {
				prio1 := t.Args[0].(Compound).Prio
				if prio1 == 0 {
					prio1 = 7
				}
				if t.Args[1].Type() == CompoundType {
					prio2 := t.Args[1].(Compound).Prio
					if prio2 == 0 {
						prio2 = 7
					}
					switch {
					case prio1 < prio && prio2 < prio:
						return "(" + t.Args[0].String() + ") " + f + " (" + t.Args[1].String() + ")"
					case prio1 < prio:
						return "(" + t.Args[0].String() + ") " + f + " " + t.Args[1].String()
					case prio2 < prio:
						return t.Args[0].String() + " " + f + " (" + t.Args[1].String() + ")"
					default:
						return t.Args[0].String() + f + t.Args[1].String()
					}
				} else {
					if prio1 < prio {
						return "(" + t.Args[0].String() + ") " + f + " " + t.Args[1].String()
					} else {
						return t.Args[0].String() + f + t.Args[1].String()
					}
				}
			} else if t.Args[1].Type() == CompoundType && t.Args[1].(Compound).Prio != 0 && t.Args[1].(Compound).Prio < prio {
				return t.Args[0].String() + " " + f + " (" + t.Args[1].String() + ")"
			}
			return t.Args[0].String() + f + t.Args[1].String()

		}
	}
	// Prio == 0
	if t.Functor == "|" {
		args := []string{}
		var oldarg Term = nil
		for _, arg := range t.Args {
			if oldarg != nil {
				args = append(args, oldarg.String())
			}
			oldarg = arg
		}
		return "[" + strings.Join(args, ",") + " | " + oldarg.String() + "]"
	}
	if t.Arity() == 0 {
		if t.HasArgs {
			return t.Functor + "()"
		} else {
			return t.Functor
		}
	}
	args := []string{}
	for _, arg := range t.Args {
		args = append(args, arg.String())
	}
	return t.Functor + "(" + strings.Join(args, ",") + ")"
}

func (t List) String() string {
	var v Term = nil
	args := []string{}

	for _, arg := range t {
		if arg.Type() == CompoundType && arg.(Compound).Functor == "|" {
			v = arg.(Compound).Args[0]
		} else {
			args = append(args, arg.String())
		}
	}
	if v != nil {
		return "[" + strings.Join(args, ", ") + " | " + v.String() + "]"
	}
	return "[" + strings.Join(args, ", ") + "]"
}

func (v Variable) String() string {
	if v.index == nil || v.index.Cmp(big.NewInt(0)) == 0 {
		return v.Name
	} else {
		return v.Name + v.index.String()
	}
}

func (t Atom) OccurVars() Vars {
	return nil
}

func (t Bool) OccurVars() Vars {
	return nil
}

func (t Int) OccurVars() Vars {
	return nil
}

func (t Float) OccurVars() Vars {
	return nil
}

func (t String) OccurVars() Vars {
	return nil
}

func (t Compound) OccurVars() Vars {
	if t.identifyOccurVars {
		return t.occurVars
	}
	occur := Vars{}
	for _, t2 := range t.Args {
		occur = append(occur, t2.OccurVars()...)
	}
	t.occurVars = occur
	t.identifyOccurVars = true
	return t.occurVars
}

func (t List) OccurVars() Vars {
	occur := Vars{}
	for _, t2 := range t {
		occur = append(occur, t2.OccurVars()...)
	}
	return occur
}

func (t Variable) OccurVars() Vars {
	return Vars{t}
}

func (t Compound) Arity() int {
	return len(t.Args)
}

// stream of pointers to big integers for renaming variables
var Counter <-chan *big.Int

// var Counter <-chan *big.Int

var InitRenamingVariables func()

func init() {
	c := make(chan *big.Int)
	reset := make(chan bool)
	i := big.NewInt(1)
	one := big.NewInt(1)
	go func() {
		for {
			select {
			case c <- i:
				i = new(big.Int).Add(i, one)
			case <-reset:
				i = one
			}
		}
	}()
	InitRenamingVariables = func() { reset <- true }
	Counter = c
}

func (v Variable) Rename() Variable {
	return Variable{Name: v.Name, index: <-Counter}
}

func Equal(t1, t2 Term) bool {
	if t1.Type() != t2.Type() {
		return false
	}
	switch t1.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return t1 == t2
	case CompoundType:
		//		fmt.Printf("## t1-Functor %s(%d), t2-Functor %s(%d)\n ", t1.(Compound).Functor, t1.(Compound).Arity(),
		//			t2.(Compound).Functor, t2.(Compound).Arity())

		if t1.(Compound).Functor != t2.(Compound).Functor ||
			t1.(Compound).Arity() != t2.(Compound).Arity() {
			//		if t1.(Compound).Prio != 3 && t2.(Compound).Prio != 3 { return false }
			// 	return EqualCompare(t1.(Compound).Functor, )
			//			fmt.Printf("## ## Functor!=Functor %v, Arity != Arity %v\n", t1.(Compound).Functor != t2.(Compound).Functor,
			//				t1.(Compound).Arity() != t2.(Compound).Arity())
			return false
		}
		for i, _ := range t1.(Compound).Args {
			if !Equal(t1.(Compound).Args[i], t2.(Compound).Args[i]) {
				//				fmt.Printf("### Arg[%v]: %s != Arg[%v]: %s \n", i, t1.(Compound).Args[i], i, t2.(Compound).Args[i])
				return false
			}
		}
		return true
	case ListType:
		if len(t1.(List)) != len(t2.(List)) {
			return false
		}
		for i, _ := range t1.(List) {
			if !Equal(t1.(List)[i], t2.(List)[i]) {
				return false
			}
		}
		return true
	case VariableType:
		if t1.(Variable).Name == t2.(Variable).Name &&
			(t1.(Variable).index.Cmp(t2.(Variable).index) == 0 ||
				(t1.(Variable).index == nil && t2.(Variable).index == nil)) {
			return true
		}
		//		fmt.Printf("## ## ## t1-name: %s, t2-name: %s, t1-idx: %v, t2-idx: %v\n",
		//			t1.(Variable).Name, t2.(Variable).Name, t1.(Variable).index, t2.(Variable).index)
		return false
	default:
		return false
	}
}

/*func copyBindings(env Bindings) Bindings {
	result := make(Bindings)
	for v, t := range env {
		result[v] = t
	}
	return result
} */

// Match updates the bindings only if the match
// is successful, in which case true is returned.
// One way match, not unification:  variables
// in t1 are bound to terms in t2.
//func Match(t1, t2 Term, env Bindings) (ok bool) {
//	ok, _ = Match1(t1, t2, env)
//	return ok
//}

func Match(t1, t2 Term, env Bindings) (env2 Bindings, ok bool) {
	if t1.Type() != VariableType && t1.Type() != t2.Type() {
		return env, false
	}
	switch t1.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return env, Equal(t1, t2)
	case CompoundType:
		if t1.(Compound).Functor != t2.(Compound).Functor ||
			t1.(Compound).Arity() != t2.(Compound).Arity() {
			return env, false
		}
		env2 := env
		for i, _ := range t1.(Compound).Args {
			env2, ok = Match(t1.(Compound).Args[i], t2.(Compound).Args[i], env2)
			if !ok {
				return env, false
			}
		}
		// update env with the new bindings
		env = env2
		/*		for v, t := range env2 {
				env[v] = t
			} */
		return env, true
	case ListType:
		lent1 := len(t1.(List))
		lent2 := len(t2.(List))
		if lent1 == 0 {
			if lent2 == 0 {
				return env, true
			} else {
				return env, false
			}
		}
		lent1m1 := lent1 - 1
		last := t1.(List)[lent1m1]
		if last.Type() == CompoundType && last.(Compound).Functor == "|" {
			if lent2 < lent1m1 {
				return env, false
			}
			env2 := env
			for i := 0; i < lent1m1; i++ {
				env2, ok = Match(t1.(List)[i], t2.(List)[i], env2)
				if !ok {
					return env, false
				}
			}
			v := last.(Compound).Args[0]
			if lent2 == lent1m1 {
				env2, ok = Match(v, List{}, env2)
			} else {
				env2, ok = Match(v, t2.(List)[lent1m1:], env2)
			}
			if !ok {
				return env, false
			}
			env = env2
			return env, true
		}
		if lent1 != lent2 {
			return env, false
		}
		env2 := env
		// for i, _ := range t1.(List) {
		for i := 0; i < lent1; i++ {
			env2, ok = Match(t1.(List)[i], t2.(List)[i], env2)
			if !ok {
				return env, false
			}
		}
		// update env with the new bindings
		env = env2
		/*	for v, t := range env2 {
			env[v] = t
		} */
		return env, true

	case VariableType:
		t3, ok := GetBinding(t1.(Variable), env)
		if !ok { // variable was not yet bound in env
			env = AddBinding(t1.(Variable), t2, env)
			return env, true
		} else {
			// return true only if the two instances of the variable
			// would be bound to the same term
			if Equal(t2, t3) {
				return env, true
			} else {
				return env, false
			}
		}
	default:
		return env, false
	}
}

// Unify two terms in equals t1 == t2
func Unify(t1, t2 Term, env Bindings) (env2 Bindings, ok bool) {
	return Unify1(t1, t2, Vars{}, env)
}

func Unify1(t1, t2 Term, visited Vars, env Bindings) (env2 Bindings, ok bool) {
	t1Type := t1.Type()
	for t1Type == VariableType {
		visited = append(visited, t1.(Variable))
		t3, ok := GetBinding(t1.(Variable), env)
		if ok {
			t1 = t3
			t1Type = t1.Type()
		} else {
			break
		}
	}

	t2Type := t2.Type()
	for t2Type == VariableType {
		visited = append(visited, t2.(Variable))
		t3, ok := GetBinding(t2.(Variable), env)
		if ok {
			t2 = t3
			t2Type = t2.Type()
		} else {
			break
		}
	}
	if t1Type == VariableType {
		if t2Type == VariableType {
			if t1.(Variable).Name == t2.(Variable).Name &&
				(t1.(Variable).index.Cmp(t2.(Variable).index) == 0 ||
					(t1.(Variable).index == nil && t2.(Variable).index == nil)) {
				// Var == Var
				return env, true
			} else {
				env2 = AddBinding(t1.(Variable), t2, env)
				return env2, true
			}
		}
		if checkOccur(visited, t2, env) {
			return nil, false
		}
		env2 = AddBinding(t1.(Variable), t2, env)
		return env2, true
	}
	if t2Type == VariableType {
		if checkOccur(visited, t1, env) {
			return nil, false
		}
		env2 = AddBinding(t2.(Variable), t1, env)
		return env2, true
	}
	if t1Type != t2Type {
		return env, false
	}
	switch t1.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return env, Equal(t1, t2)
	case CompoundType:
		if t1.(Compound).Functor != t2.(Compound).Functor ||
			t1.(Compound).Arity() != t2.(Compound).Arity() {
			return env, false
		}
		env2 := env
		for i, _ := range t1.(Compound).Args {
			env2, ok = Unify1(t1.(Compound).Args[i], t2.(Compound).Args[i], visited, env2)
			if !ok {
				return env, false
			}
		}
		// update env with the new bindings
		env = env2
		/*		for v, t := range env2 {
				env[v] = t
			} */
		return env, true
	case ListType:
		lent1 := len(t1.(List))
		lent2 := len(t2.(List))
		if lent1 == 0 {
			if lent2 == 0 {
				return env, true
			} else {
				return env, false
			}
		}
		lent1m1 := lent1 - 1
		last := t1.(List)[lent1m1]
		if last.Type() == CompoundType && last.(Compound).Functor == "|" {
			if lent2 < lent1m1 {
				return env, false
			}
			env2 := env
			for i := 0; i < lent1m1; i++ {
				env2, ok = Unify1(t1.(List)[i], t2.(List)[i], visited, env2)
				if !ok {
					return env, false
				}
			}
			v := last.(Compound).Args[0]
			if lent2 == lent1m1 {
				env2, ok = Unify1(v, List{}, visited, env2)
			} else {
				env2, ok = Unify1(v, t2.(List)[lent1m1:], visited, env2)
			}
			if !ok {
				return env, false
			}
			env = env2
			return env, true
		}
		if lent1 != lent2 {
			return env, false
		}
		env2 := env
		// for i, _ := range t1.(List) {
		for i := 0; i < lent1; i++ {
			env2, ok = Unify1(t1.(List)[i], t2.(List)[i], visited, env2)
			if !ok {
				return env, false
			}
		}
		// update env with the new bindings
		env = env2
		/*	for v, t := range env2 {
			env[v] = t
		} */
		return env, true
	default:
		return env, false
	}
}

func checkOccur(v Vars, t Term, env Bindings) bool {

	for _, termv := range t.OccurVars() {
		// fmt.Printf("** Var %s in Term: %s \n", termv, t)
		for _, visitv := range v {
			if termv.Name == visitv.Name && termv.index.Cmp(visitv.index) == 0 {
				return true
			}
		}
		t2, ok := GetBinding(termv, env)
		if ok {
			for _, termv := range t2.OccurVars() {
				for _, visitv := range v {
					if termv.Name == visitv.Name && termv.index.Cmp(visitv.index) == 0 {
						return true
					}
				}
			}
		}

	}
	return false
}

func Arity(t Term) int {
	if t.Type() != CompoundType {
		return 0
	}
	return t.(Compound).Arity()
}

func isTriple(t Term) bool {
	return Arity(t) == 2
}

func Functor(t Term) (result string, ok bool) {
	switch t.Type() {
	case AtomType:
		return t.String(), true
	case CompoundType:
		return t.(Compound).Functor, true
	default:
		return result, false
	}
}

// Predicate is a synonym for Functor
func Predicate(t Term) (string, bool) {
	return Functor(t)
}

func Subject(t Term) (result Term, ok bool) {
	if isTriple(t) {
		return t.(Compound).Args[0], true
	}
	return result, false
}

func Object(t Term) (result Term, ok bool) {
	if isTriple(t) {
		return t.(Compound).Args[1], true
	}
	return result, false
}

// Substitute: replace variables in the term t with
// their bindings in the env, if they are bound.
// Follows variable chains, so that if a variable
// is bound to a variable, the second variable is also
// substituted if it is bound in env, recursively.
func Substitute(t Term, env Bindings) Term {
	return Substitute1(t, map[string]bool{}, env)
}

func Substitute1(t Term, visited map[string]bool, env Bindings) Term {

	switch t.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return t
	case CompoundType:
		args := []Term{}
		for _, t2 := range t.(Compound).Args {
			args = append(args, Substitute(t2, env))
		}
		return Compound{Functor: t.(Compound).Functor, Id: t.(Compound).Id,
			Prio: t.(Compound).Prio, Args: args}
	case ListType:
		l := []Term{}
		for _, t2 := range t.(List) {
			l = append(l, Substitute(t2, env))
		}
		return List(l)
	case VariableType:
		result := t
		visited[fmt.Sprintf("%s#%v", t.(Variable).Name, t.(Variable).index)] = true
		t2, ok := GetBinding(t.(Variable), env)
		for ok == true {
			result = t2
			if t2.Type() == VariableType && !visited[fmt.Sprintf("%s#%v", t2.(Variable).Name, t2.(Variable).index)] {
				t2, ok = GetBinding(t2.(Variable), env)
				continue
			} else {
				if t2.Type() != VariableType {
					result = Substitute1(t2, visited, env)
				}
				break
			}
		}
		return result
	default:
		return t
	}
}

// Substitute: replace variables in the term t with
// their bindings in the Build-In environment (BIVarEqTerm),
// if they are bound.
// Follows variable chains, so that if a variable
// is bound to a variable, the second variable is also
// substituted if it is bound in env, recursively.
func SubstituteBiEnv(t Term, biEnv Bindings) (Term, bool) {
	ok := false
	visited := map[Variable]bool{}

	switch t.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return t, ok
	case CompoundType:
		args := []Term{}
		for _, t2 := range t.(Compound).Args {
			a, ok2 := SubstituteBiEnv(t2, biEnv)
			ok = ok || ok2
			args = append(args, a)
		}
		return Compound{Functor: t.(Compound).Functor, Id: t.(Compound).Id,
			Prio: t.(Compound).Prio, Args: args}, ok
	case ListType:
		l := []Term{}
		for _, t2 := range t.(List) {
			l1, ok2 := SubstituteBiEnv(t2, biEnv)
			ok = ok || ok2
			l = append(l, l1)
		}
		return List(l), ok
	case VariableType:

		t2, ok2 := GetBinding(t.(Variable), biEnv)
		ok = ok || ok2

		for ok2 == true {
			t = t2
			if t2.Type() == VariableType && !visited[t2.(Variable)] {
				t2, ok2 = GetBinding(t2.(Variable), biEnv)
				continue
			} else {
				break
			}
		}
		return t, ok
	default:
		return t, ok
	}
}

// Substitute: replace variables in the term t with
// their bindings in the env or in the Build-In environment (BIVarEqTerm),
// if they are bound. If their are not bound rename
// the body-varaible of the rule (very late renaming).
// Follows variable chains, so that if a variable
// is bound to a variable, the second variable is also
// substituted if it is bound in env, recursively.
func RenameAndSubstitute(t Term, idx *big.Int, env Bindings) Term {
	// visited := map[Variable]bool{}

	switch t.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return t
	case CompoundType:
		args := []Term{}
		for _, t2 := range t.(Compound).Args {
			args = append(args, RenameAndSubstitute(t2, idx, env))
		}
		return Compound{Functor: t.(Compound).Functor, Id: t.(Compound).Id,
			Prio: t.(Compound).Prio, Args: args}
	case ListType:
		l := []Term{}
		for _, t2 := range t.(List) {
			l = append(l, RenameAndSubstitute(t2, idx, env))
		}
		return List(l)
	case VariableType:

		t2, ok := GetBinding(t.(Variable), env)
		if !ok {
			// very late variable renaming
			t = Variable{Name: t.(Variable).Name, index: idx}
			// visited[t.(Variable)] = true
			t2, ok = GetBinding(t.(Variable), env)
			if !ok {
				return t
			}
			t = t2
			return t
		}
		//		for ok == true {
		//			t = t2
		//			if t2.Type() == VariableType && !visited[t2.(Variable)] {
		//				visited[t2.(Variable)] = true
		//				t2, ok = GetBinding(t.(Variable), env)
		//			} else {
		//				break
		//			}
		//		}
		t = t2
		return t
	default:
		return t
	}
}
