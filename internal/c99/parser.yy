%{
// Copyright 2017 The C99 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Based on [0], 6.5-6.10. Substantial portions of expression AST size
// optimizations are from [1], license of which follows.
//
// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf
// [1]: https://github.com/rsc/c2go/blob/fc8cbfad5a47373828c81c7a56cccab8b221d310/cc/cc.y

// ----------------------------------------------------------------------------

// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This grammar is derived from the C grammar in the 'ansitize'
// program, which carried this notice:
// 
// Copyright (c) 2006 Russ Cox, 
// 	Massachusetts Institute of Technology
// 
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated
// documentation files (the "Software"), to deal in the
// Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute,
// sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
// 
// The above copyright notice and this permission notice shall
// be included in all copies or substantial portions of the
// Software.
// 
// The software is provided "as is", without warranty of any
// kind, express or implied, including but not limited to the
// warranties of merchantability, fitness for a particular
// purpose and noninfringement.  In no event shall the authors
// or copyright holders be liable for any claim, damages or
// other liability, whether in an action of contract, tort or
// otherwise, arising from, out of or in connection with the
// software or the use or other dealings in the software.

package c99

import (
	"github.com/cznic/xc"
)
%}

%union {
	Token			xc.Token // Must be exported
	node			Node
}

%token
	/*yy:token "'%c'"            */ CHARCONST		"character constant"
	/*yy:token "1.%d"            */ FLOATCONST		"floating-point constant"
	/*yy:token "%c"              */ IDENTIFIER		"identifier"
	/*yy:token "%d"              */ INTCONST		"integer constant"
	/*yy:token "L'%c'"           */ LONGCHARCONST		"long character constant"
	/*yy:token "L\"%c\""         */ LONGSTRINGLITERAL	"long string constant"
	/*yy:token "%d"              */ PPNUMBER		"preprocessing number"
	/*yy:token "\"%c\""          */ STRINGLITERAL		"string literal"

	/*yy:token "\U00100000"      */	CONSTANT_EXPRESSION	1048576	"constant expression prefix"	// 0x100000 = 1048576
	/*yy:token "\U00100001"      */	TRANSLATION_UNIT	1048577	"translation unit prefix"

	ADDASSIGN			"+="
	ANDAND				"&&"
	ANDASSIGN			"&="
	ARROW				"->"
	AUTO				"auto"
	BOOL				"_Bool"
	BREAK				"break"
	CASE				"case"
	CHAR				"char"
	COMPLEX				"_Complex"
	CONST				"const"
	CONTINUE			"continue"
	DDD				"..."
	DEC				"--"
	DEFAULT				"default"
	DIRECTIVE			// # when it's the first non white token on a line.
	DIVASSIGN			"/="
	DO				"do"
	DOUBLE				"double"
	ELSE				"else"
	ENUM				"enum"
	EQ				"=="
	EXTERN				"extern"
	FLOAT				"float"
	FOR				"for"
	GEQ				">="
	GOTO				"goto"
	NON_REPL			// [0]6.10.3.4-2
	IF				"if"
	INC				"++"
	INLINE				"inline"
	INT				"int"
	LEQ				"<="
	LONG				"long"
	LSH				"<<"
	LSHASSIGN			"<<="
	MODASSIGN			"%="
	MULASSIGN			"*="
	NEQ				"!="
	ORASSIGN			"|="
	OROR				"||"
	PPPASTE				"##"
	REGISTER			"register"
	RESTRICT			"restrict"
	RETURN				"return"
	RSH				">>"
	RSHASSIGN			">>="
	SENTINEL			// Hide set marker.
	SHORT				"short"
	SIGNED				"signed"
	SIZEOF				"sizeof"
	STATIC				"static"
	STRUCT				"struct"
	SUBASSIGN			"-="
	SWITCH				"switch"
	TYPEDEF				"typedef"
	TYPEDEF_NAME			"typedef name"
	TYPEOF				"typeof"
	UNION				"union"
	UNSIGNED			"unsigned"
	VOID				"void"
	VOLATILE			"volatile"
	WHILE				"while"
	XORASSIGN			"^="

%type	<node>
	AbstractDeclarator		"abstract declarator"
	AbstractDeclaratorOpt		"optional abstract declarator"
	ArgumentExprList		"argument expression list"
	ArgumentExprListOpt		"optional argument expression list"
	BlockItem			"block item"
	BlockItemList			"block item list"
	BlockItemListOpt		"optional block item list"
	CommaOpt			"optional comma"
	CompoundStmt			"compound statement"
	ConstExpr			"constant expression"
	Declaration			"declaration"
	DeclarationList			"declaration list"
	DeclarationListOpt		"optional declaration list"
	DeclarationSpecifiers		"declaration specifiers"
	DeclarationSpecifiersOpt	"optional declaration specifiers"
	Declarator			"declarator"
	DeclaratorOpt			"optional declarator"
	Designation			"designation"
	Designator			"designator"
	DesignatorList			"designator list"
	DirectAbstractDeclarator	"direct abstract declarator"
	DirectAbstractDeclaratorOpt	"optional direct abstract declarator"
	DirectDeclarator		"direct declarator"
	EnumSpecifier			"enum specifier"
	EnumerationConstant		"enumearation constant"
	Enumerator			"enumerator"
	EnumeratorList			"enumerator list"
	Expr				"expression"
	ExprList			"expression list"
	ExprListOpt			"optional expression list"
	ExprOpt				"optional expression"
	ExprStmt			"expression statement"
	ExternalDeclaration		"external declaration"
	FunctionBody			"function body"
	FunctionDefinition		"function definition"
	FunctionSpecifier		"function specifier"
	IdentifierList			"identifier list"
	IdentifierListOpt		"optional identifier list"
	IdentifierOpt			"optional identifier"
	InitDeclarator			"init declarator"
	InitDeclaratorList		"init declarator list"
	InitDeclaratorListOpt		"optional init declarator list"
	Initializer			"initializer"
	InitializerList			"initializer list"
	IterationStmt			"iteration statement"
	JumpStmt			"jump statement"
	LabeledStmt			"labeled statement"
	ParameterDeclaration		"parameter declaration"
	ParameterList			"parameter list"
	ParameterTypeList		"parameter type list"
	ParameterTypeListOpt		"optional parameter type list"
	Pointer				"pointer"
	PointerOpt			"optional pointer"
	SelectionStmt			"selection statement"
	SpecifierQualifierList		"specifier qualifier list"
	SpecifierQualifierListOpt	"optional specifier qualifier list"
	Stmt				"statement"
	StorageClassSpecifier		"storage class specifier"
	StructDeclaration		"struct declaration"
	StructDeclarationList		"struct declaration list"
	StructDeclarator		"struct declarator"
	StructDeclaratorList		"struct declarator list"
	StructOrUnion			"struct-or-union"
	StructOrUnionSpecifier		"struct-or-union specifier"
	TranslationUnit			"translation unit"
	TypeName			"type name"
	TypeQualifier			"type qualifier"
	TypeQualifierList		"type qualifier list"
	TypeQualifierListOpt		"optional type qualifier list"
	TypeSpecifier			"type specifier"
	VolatileOpt			"optional volatile"


%precedence	NOSEMI
%precedence	';'

%precedence	NOELSE
%precedence	ELSE

%right		'=' ADDASSIGN ANDASSIGN DIVASSIGN LSHASSIGN MODASSIGN MULASSIGN
		ORASSIGN RSHASSIGN SUBASSIGN XORASSIGN
		
%right		':' '?'
%left		OROR
%left		ANDAND
%left		'|'
%left		'^'
%left		'&'
%left		EQ NEQ
%left		'<' '>' GEQ LEQ
%left		LSH RSH
%left		'+' '-' 
%left		'%' '*' '/'
%precedence	CAST
%left		'!' '~' SIZEOF UNARY
%right		'(' '.' '[' ARROW DEC INC

%%

                        /*yy:ignore */
                        Start:
                        	CONSTANT_EXPRESSION ConstExpr
				{
					lx.ast = $2
				}
                        |	TRANSLATION_UNIT TranslationUnit
				{
					lx.ast = $2
				}

                        // [0]6.4.4.3
                        EnumerationConstant:
                        	IDENTIFIER

                        // [0]6.5.2
                        ArgumentExprList:
                        	Expr
                        |	ArgumentExprList ',' Expr

                        ArgumentExprListOpt:
                        	/* empty */ {}
                        |	ArgumentExprList

                        // [0]6.5.16
			//yy:field	Value	*Value
/*yy:case PreInc  */ Expr:
                        	"++" Expr
/*yy:case PreDec     */ |	"--" Expr
/*yy:case SizeOfType */ |	"sizeof" '(' TypeName ')' %prec SIZEOF
/*yy:case SizeofExpr */ |	"sizeof" Expr
/*yy:case Not        */ |	'!' Expr
/*yy:case Addrof     */ |	'&' Expr %prec UNARY
/*yy:case PExprList  */ |	'(' ExprList ')'
/*yy:case CompLit    */ |	'(' TypeName ')' '{' InitializerList CommaOpt '}'
/*yy:case Cast       */ |	'(' TypeName ')' Expr %prec CAST
/*yy:case Deref      */ |	'*' Expr %prec UNARY
/*yy:case UnaryPlus  */ |	'+' Expr %prec UNARY
/*yy:case UnaryMinus */ |	'-' Expr %prec UNARY
/*yy:case Cpl        */ |	'~' Expr
/*yy:case Char       */ |	CHARCONST
/*yy:case Ne         */ |	Expr "!=" Expr
/*yy:case ModAssign  */ |	Expr "%=" Expr
/*yy:case LAnd       */ |	Expr "&&" Expr
/*yy:case AndAssign  */ |	Expr "&=" Expr
/*yy:case MulAssign  */ |	Expr "*=" Expr
/*yy:case PostInt    */ |	Expr "++"
/*yy:case AddAssign  */ |	Expr "+=" Expr
/*yy:case PostDec    */ |	Expr "--"
/*yy:case SubAssign  */ |	Expr "-=" Expr
/*yy:case PSelect    */ |	Expr "->" IDENTIFIER
/*yy:case DivAssign  */ |	Expr "/=" Expr
/*yy:case Lsh        */ |	Expr "<<" Expr
/*yy:case LshAssign  */ |	Expr "<<=" Expr
/*yy:case Le         */ |	Expr "<=" Expr
/*yy:case Eq         */ |	Expr "==" Expr
/*yy:case Ge         */ |	Expr ">=" Expr
/*yy:case Rsh        */ |	Expr ">>" Expr
/*yy:case RshAssign  */ |	Expr ">>=" Expr
/*yy:case XorAssign  */ |	Expr "^=" Expr
/*yy:case OrAssign   */ |	Expr "|=" Expr
/*yy:case LOr        */ |	Expr "||" Expr
/*yy:case Mod        */ |	Expr '%' Expr
/*yy:case And        */ |	Expr '&' Expr
/*yy:case Call       */ |	Expr '(' ArgumentExprListOpt ')'
/*yy:case Mul        */ |	Expr '*' Expr
/*yy:case Add        */ |	Expr '+' Expr
/*yy:case Sub        */ |	Expr '-' Expr
/*yy:case Select     */ |	Expr '.' IDENTIFIER
/*yy:case Div        */ |	Expr '/' Expr
/*yy:case Lt         */ |	Expr '<' Expr
/*yy:case Assign     */ |	Expr '=' Expr
/*yy:case Gt         */ |	Expr '>' Expr
/*yy:case Cond       */ |	Expr '?' ExprList ':' Expr
/*yy:case Index      */ |	Expr '[' ExprList ']'
/*yy:case Xor        */ |	Expr '^' Expr
/*yy:case Or         */ |	Expr '|' Expr
/*yy:case Float      */ |	FLOATCONST
/*yy:case Ident      */ |	IDENTIFIER %prec NOSEMI
/*yy:case Int        */ |	INTCONST
/*yy:case LChar      */ |	LONGCHARCONST
/*yy:case LString    */ |	LONGSTRINGLITERAL
/*yy:case String     */ |	STRINGLITERAL

                        ExprOpt:
                        	/* empty */ {}
                        |	Expr

                        // [0]6.5.17
                        //yy:list
			//yy:field	Value	*Value
                        ExprList:
                        	Expr
                        |	ExprList ',' Expr

                        ExprListOpt:
                        	/* empty */ {}
                        |	ExprList

                        // [0]6.6
			//yy:field	Value	*Value
                        ConstExpr:
                        	Expr

                        // [0]6.7
			Declaration:
                        	DeclarationSpecifiers InitDeclaratorListOpt ';'

                        // [0]6.7
/*yy:case Func       */ DeclarationSpecifiers:
                        	FunctionSpecifier DeclarationSpecifiersOpt
/*yy:case Strorage   */ |	StorageClassSpecifier DeclarationSpecifiersOpt
/*yy:case Qualifier  */ |	TypeQualifier DeclarationSpecifiersOpt
/*yy:case Specifier  */ |	TypeSpecifier DeclarationSpecifiersOpt

                        DeclarationSpecifiersOpt:
                        	/* empty */ {}
                        |	DeclarationSpecifiers

                        // [0]6.7
                        InitDeclaratorList:
                        	InitDeclarator
                        |	InitDeclaratorList ',' InitDeclarator

                        InitDeclaratorListOpt:
                        	/* empty */ {}
                        |	InitDeclaratorList

                        // [0]6.7
/*yy:case Base       */ InitDeclarator:
                        	Declarator
/*yy:case Init       */ |	Declarator '=' Initializer

                        // [0]6.7.1
/*yy:case Auto       */ StorageClassSpecifier:
                        	"auto"
/*yy:case Extern     */ |	"extern"
/*yy:case Register   */ |	"register"
/*yy:case Static     */ |	"static"
/*yy:case Typedef    */ |	"typedef"

                        // [0]6.7.2
/*yy:case Bool       */ TypeSpecifier:
                        	"_Bool"
/*yy:case Complex    */ |	"_Complex"
/*yy:case Char       */ |	"char"
/*yy:case Double     */ |	"double"
/*yy:case Float      */ |	"float"
/*yy:case Int        */ |	"int"
/*yy:case Long       */ |	"long"
/*yy:case Short      */ |	"short"
/*yy:case Signed     */ |	"signed"
/*yy:case Unsigned   */ |	"unsigned"
/*yy:case Void       */ |	"void"
/*yy:case Enum       */ |	EnumSpecifier
/*yy:case Struct     */ |	StructOrUnionSpecifier
//yy:example "\U00100001 typedef int foo; foo bar;"
/*yy:case Name       */ |	TYPEDEF_NAME

                        // [0]6.7.2.1
/*yy:case Tag        */ StructOrUnionSpecifier:
                        	StructOrUnion IDENTIFIER
/*yy:case Define     */ |	StructOrUnion IdentifierOpt '{' StructDeclarationList '}'

                        // [0]6.7.2.1
/*yy:case Struct     */ StructOrUnion:
                        	"struct"
/*yy:case Union      */ |	"union"

                        // [0]6.7.2.1
                        StructDeclarationList:
                        	StructDeclaration
                        |	StructDeclarationList StructDeclaration

                        // [0]6.7.2.1
                        StructDeclaration:
				SpecifierQualifierList StructDeclaratorList ';'

                        // [0]6.7.2.1
/*yy:case Qualifier  */ SpecifierQualifierList:
                        	TypeQualifier SpecifierQualifierListOpt
/*yy:case Specifier  */ |	TypeSpecifier SpecifierQualifierListOpt

                        SpecifierQualifierListOpt:
                        	/* empty */ {}
                        |	SpecifierQualifierList

                        // [0]6.7.2.1
                        StructDeclaratorList:
                        	StructDeclarator
                        |	StructDeclaratorList ',' StructDeclarator

                        // [0]6.7.2.1
/*yy:case Base       */ StructDeclarator:
                        	Declarator
/*yy:case Bits       */ |	DeclaratorOpt ':' ConstExpr

                        CommaOpt:
                        	/* empty */ {}
                        |	','

                        // [0]6.7.2.2
/*yy:case Tag        */ EnumSpecifier:
                        	"enum" IDENTIFIER
/*yy:case Define     */ |	"enum" IdentifierOpt '{' EnumeratorList  CommaOpt '}'

                        // [0]6.7.2.2
                        EnumeratorList:
                        	Enumerator
                        |	EnumeratorList ',' Enumerator

                        // [0]6.7.2.2
/*yy:case Base       */ Enumerator:
                        	EnumerationConstant
/*yy:case Init       */ |	EnumerationConstant '=' ConstExpr

                        // [0]6.7.3
/*yy:case Const      */ TypeQualifier:
                        	"const"
/*yy:case Restrict   */ |	"restrict"
/*yy:case Volatile   */ |	"volatile"

                        // [0]6.7.4
			FunctionSpecifier:
				"inline"

                        // [0]6.7.5
                        Declarator:
                        	PointerOpt DirectDeclarator

                        DeclaratorOpt:
                        	/* empty */ {}
                        |	Declarator

                        // [0]6.7.5
/*yy:case Paren      */ DirectDeclarator:
                        	'(' Declarator ')'
/*yy:case IdentList  */ |	DirectDeclarator '(' IdentifierListOpt ')'
/*yy:case ParamList  */ |	DirectDeclarator '(' ParameterTypeList ')'
/*yy:case ArraySize  */ |	DirectDeclarator '[' "static" TypeQualifierListOpt Expr ']'
/*yy:case ArraySize2 */ |	DirectDeclarator '[' TypeQualifierList "static" Expr ']'
/*yy:case ArrayVar   */ |	DirectDeclarator '[' TypeQualifierListOpt '*' ']'
/*yy:case Array      */ |	DirectDeclarator '[' TypeQualifierListOpt ExprOpt ']'
/*yy:case Ident      */ |	IDENTIFIER

                        // [0]6.7.5
/*yy:case Base       */ Pointer:
                        	'*' TypeQualifierListOpt
/*yy:case Ptr        */ |	'*' TypeQualifierListOpt Pointer

                        PointerOpt:
                        	/* empty */ {}
                        |	Pointer

                        // [0]6.7.5
                        TypeQualifierList:
                        	TypeQualifier
                        |	TypeQualifierList TypeQualifier

                        TypeQualifierListOpt:
                        	/* empty */ {}
                        |	TypeQualifierList

                        // [0]6.7.5
/*yy:case Base       */ ParameterTypeList:
                        	ParameterList
/*yy:case Dots       */ |	ParameterList ',' "..."

                        ParameterTypeListOpt:
                        	/* empty */ {}
                        |	ParameterTypeList

                        // [0]6.7.5
                        ParameterList:
                        	ParameterDeclaration
                        |	ParameterList ',' ParameterDeclaration

                        // [0]6.7.5
/*yy:case Abstract   */ ParameterDeclaration:
                        	DeclarationSpecifiers AbstractDeclaratorOpt
/*yy:case Declarator */ |	DeclarationSpecifiers Declarator

                        // [0]6.7.5
                        IdentifierList:
                        	IDENTIFIER
                        |	IdentifierList ',' IDENTIFIER

                        IdentifierListOpt:
                        	/* empty */ {}
                        |	IdentifierList

                        IdentifierOpt:
                        	/* empty */ {}
                        |	IDENTIFIER

                        // [0]6.7.6
                        TypeName:
                        	SpecifierQualifierList AbstractDeclaratorOpt

                        // [0]6.7.6
/*yy:case Pointer    */ AbstractDeclarator:
                        	Pointer
/*yy:case Abstract   */ |	PointerOpt DirectAbstractDeclarator

                        AbstractDeclaratorOpt:
                        	/* empty */ {}
                        |	AbstractDeclarator

                        // [0]6.7.6
/*yy:case Abstract   */ DirectAbstractDeclarator:
                        	'(' AbstractDeclarator ')'
/*yy:case ParamList  */ |	'(' ParameterTypeListOpt ')'
/*yy:case DFn        */ |	DirectAbstractDeclarator '(' ParameterTypeListOpt ')'
/*yy:case DArrSize   */ |	DirectAbstractDeclaratorOpt '[' "static" TypeQualifierListOpt Expr ']'
/*yy:case DArrVL     */ |	DirectAbstractDeclaratorOpt '[' '*' ']'
/*yy:case DArr       */ |	DirectAbstractDeclaratorOpt '[' ExprOpt ']'
/*yy:case DArrSize2  */ |	DirectAbstractDeclaratorOpt '[' TypeQualifierList "static" Expr ']'
/*yy:case DArr2      */ |	DirectAbstractDeclaratorOpt '[' TypeQualifierList ExprOpt ']'

                        DirectAbstractDeclaratorOpt:
                        	/* empty */ {}
                        |	DirectAbstractDeclarator

                        // [0]6.7.8
/*yy:case CompLit    */ Initializer:
                        	'{' InitializerList CommaOpt '}'
/*yy:case Expr       */ |	Expr

                        // [0]6.7.8
                        InitializerList:
                        	/* empty */ {}
                        |	Initializer
                        |	Designation Initializer
                        |	InitializerList ',' Initializer
                        |	InitializerList ',' Designation Initializer

                        // [0]6.7.8
                        Designation:
                        	DesignatorList '='

                        // [0]6.7.8
                        DesignatorList:
                        	Designator
                        |	DesignatorList Designator

                        // [0]6.7.8
/*yy:case Field      */ Designator:
                        	'.' IDENTIFIER
/*yy:case Index      */ |	'[' ConstExpr ']'

                        // [0]6.8
/*yy:case Block      */ Stmt:
				CompoundStmt
/*yy:case Expr       */ |	ExprStmt
/*yy:case Iter       */ |	IterationStmt
/*yy:case Jump       */ |	JumpStmt
/*yy:case Labeled    */ |	LabeledStmt
/*yy:case Select     */ |	SelectionStmt

                        // [0]6.8.1
/*yy:case SwitchCase */ LabeledStmt:
                        	"case" ConstExpr ':' Stmt
/*yy:case Default    */ |	"default" ':' Stmt
/*yy:case Label      */ |	IDENTIFIER ':' Stmt

                        // [0]6.8.2
                        CompoundStmt:
                        	'{' BlockItemListOpt '}'

                        // [0]6.8.2
                        BlockItemList:
                        	BlockItem
                        |	BlockItemList BlockItem

                        BlockItemListOpt:
                        	/* empty */ {}
                        |	BlockItemList

                        // [0]6.8.2
/*yy:case Decl       */ BlockItem:
                        	Declaration
/*yy:case Stmt       */ |	Stmt

                        // [0]6.8.3
                        ExprStmt:
                        	ExprListOpt ';'

                        // [0]6.8.4
/*yy:case IfElse     */ SelectionStmt:
                        	"if" '(' ExprList ')' Stmt "else" Stmt
/*yy:case If         */ |	"if" '(' ExprList ')' Stmt %prec NOELSE
/*yy:case Switch     */ |	"switch" '(' ExprList ')' Stmt

                        // [0]6.8.5
/*yy:case Do         */ IterationStmt:
                        	"do" Stmt "while" '(' ExprList ')' ';'
/*yy:case ForDecl    */ |	"for" '(' Declaration ExprListOpt ';' ExprListOpt ')' Stmt
/*yy:case For        */ |	"for" '(' ExprListOpt ';' ExprListOpt ';' ExprListOpt ')' Stmt
/*yy:case While      */ |	"while" '(' ExprList ')' Stmt

                        // [0]6.8.6
/*yy:case Break      */ JumpStmt:
                        	"break" ';'
/*yy:case Continue   */ |	"continue" ';'
/*yy:case Goto       */ |	"goto" IDENTIFIER ';'
/*yy:case Return     */ |	"return" ExprListOpt ';'

                        // [0]6.9
                        //yy:list
                        TranslationUnit:
                        	ExternalDeclaration
                        |	TranslationUnit ExternalDeclaration

                        // [0]6.9
/*yy:case Decl       */ ExternalDeclaration:
				Declaration
/*yy:case Func       */ |	FunctionDefinition

                        // [0]6.9.1
			FunctionDefinition:
                        	DeclarationSpecifiers Declarator DeclarationListOpt FunctionBody

			FunctionBody:
				CompoundStmt

                        // [0]6.9.1
                        DeclarationList:
                        	Declaration
                        |	DeclarationList Declaration

                        DeclarationListOpt:
                        	/* empty */ {}
                        |	DeclarationList

                        VolatileOpt:
                        	/* empty */ {}
                        |	"volatile"
