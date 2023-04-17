package parser

import (
	"errors"
	"fmt"

	"permify/pkg/dsl/ast"
	"permify/pkg/dsl/lexer"
	"permify/pkg/dsl/token"
	"permify/pkg/dsl/utils"
)

const (
	// iota is a special identifier that is automatically set to 0 in this case, and increments by 1 for each subsequent constant declaration. By assigning the value to the blank identifier _, it is effectively ignored.
	_ int = iota

	// LOWEST precedence level for lowest precedence
	LOWEST
	// LOGIC precedence level for logical operators (AND, OR)
	LOGIC
	// PREFIX precedence level for prefix operators (NOT)
	PREFIX
)

var precedences = map[token.Type]int{ // a map that assigns precedence levels to different token types
	token.AND: LOGIC,
	token.OR:  LOGIC,
}

// Parser is a struct that contains information and functions related to parsing
type Parser struct {
	// a pointer to a Lexer object that will provide tokens for parsing
	l *lexer.Lexer
	// the current token being processed
	currentToken token.Token
	// the next token after currentToken
	peekToken token.Token
	// a slice of error messages that are generated during parsing
	errors []string
	// a map that associates prefix parsing functions with token types
	prefixParseFns map[token.Type]prefixParseFn
	// a map that associates infix parsing functions with token types
	infixParseFunc map[token.Type]infixParseFn

	// entity references
	// a map that stores entity types as keys and an empty struct as value, indicating that the entity type has been referenced
	entityReferences map[string]struct{}

	// relation references
	// a map that stores relation types as keys and a slice of relation type statements as value
	// relation types are of the form entity_type#relation_name
	relationReferences map[string][]ast.RelationTypeStatement

	// action references
	// a map that stores action types as keys and an empty struct as value, indicating that the action type has been referenced
	// action types are of the form entity_type#action_name
	actionReferences map[string]struct{}

	// relational references
	// a map that stores relational reference types as keys and a RelationalReferenceType as value
	// relational reference types are of the form entity_type#relation_name, entity_type#action_name
	relationalReferences map[string]ast.RelationalReferenceType
}

type (
	// a function that parses prefix expressions and returns an ast.Expression and error
	prefixParseFn func() (ast.Expression, error)

	// a function that parses infix expressions and returns an ast.Expression and error
	infixParseFn func(ast.Expression) (ast.Expression, error)
)

// NewParser creates a new Parser object with the given input string
func NewParser(str string) (p *Parser) {
	// initialize a new Parser object with the given input string and default values for other fields
	p = &Parser{
		l:                    lexer.NewLexer(str),                      // create a new Lexer object with the input string
		errors:               []string{},                               // initialize an empty slice of error messages
		entityReferences:     map[string]struct{}{},                    // initialize an empty map for entity references
		relationReferences:   map[string][]ast.RelationTypeStatement{}, // initialize an empty map for relation references
		actionReferences:     map[string]struct{}{},                    // initialize an empty map for action references
		relationalReferences: map[string]ast.RelationalReferenceType{}, // initialize an empty map for relational references
	}

	// register prefix parsing functions for token types IDENT and NOT
	p.prefixParseFns = make(map[token.Type]prefixParseFn) // initialize an empty map for prefix parsing functions
	p.registerPrefix(token.IDENT, p.parseIdentifier)      // associate the parseIdentifier function with the IDENT token type
	p.registerPrefix(token.NOT, p.parsePrefixExpression)  // associate the parsePrefixExpression function with the NOT token type

	// register infix parsing functions for token types AND and OR
	p.infixParseFunc = make(map[token.Type]infixParseFn) // initialize an empty map for infix parsing functions
	p.registerInfix(token.AND, p.parseInfixExpression)   // associate the parseInfixExpression function with the AND token type
	p.registerInfix(token.OR, p.parseInfixExpression)    // associate the parseInfixExpression function with the OR token type

	return // return the newly created Parser object
}

// setEntityReference adds a new entity reference to the Parser's entityReferences map
func (p *Parser) setEntityReference(key string) error {
	// if the entityReferences map is nil, initialize it
	if p.entityReferences == nil {
		p.entityReferences = map[string]struct{}{}
	}
	// check if the entity type has already been referenced, and return an error if it has
	if _, ok := p.entityReferences[key]; ok {
		p.duplicationError(key) // generate an error message indicating a duplication error
		return p.Error()        // return the error message
	}
	// add the entity type to the entityReferences map
	p.entityReferences[key] = struct{}{}
	return nil // return nil to indicate that there was no error
}

// setRelationReference adds a new relation reference to the Parser's relationReferences and relationalReferences maps
func (p *Parser) setRelationReference(key string, types []ast.RelationTypeStatement) error {
	// if the relationReferences map is nil, initialize it
	if p.relationReferences == nil {
		p.relationReferences = map[string][]ast.RelationTypeStatement{}
	}
	// check if the relation type has already been referenced, and return an error if it has
	if _, ok := p.relationReferences[key]; ok {
		p.duplicationError(key) // generate an error message indicating a duplication error
		return p.Error()        // return the error message
	}
	// check if the relation type has already been added to the relationalReferences map, and return an error if it has
	if _, ok := p.relationalReferences[key]; ok {
		p.duplicationError(key) // generate an error message indicating a duplication error
		return p.Error()        // return the error message
	}
	// add the relation type and its associated RelationTypeStatements to the relationReferences map
	p.relationReferences[key] = types
	// add the relation type to the relationalReferences map, with a value of RELATION to indicate that it is a relation reference
	p.relationalReferences[key] = ast.RELATION
	return nil // return nil to indicate that there was no error
}

// setPermissionReference adds a new action reference to the Parser's actionReferences and relationalReferences maps
func (p *Parser) setPermissionReference(key string) error {
	// if the actionReferences map is nil, initialize it
	if p.actionReferences == nil {
		p.actionReferences = map[string]struct{}{}
	}
	// check if the action type has already been referenced, and return an error if it has
	if _, ok := p.actionReferences[key]; ok {
		p.duplicationError(key) // generate an error message indicating a duplication error
		return p.Error()        // return the error message
	}
	// check if the action type has already been added to the relationalReferences map, and return an error if it has
	if _, ok := p.relationalReferences[key]; ok {
		p.duplicationError(key) // generate an error message indicating a duplication error
		return p.Error()        // return the error message
	}
	// add the action type to the actionReferences map
	p.actionReferences[key] = struct{}{}
	// add the action type to the relationalReferences map, with a value of PERMISSION to indicate that it is an action reference
	p.relationalReferences[key] = ast.PERMISSION
	return nil // return nil to indicate that there was no error
}

// next retrieves the next non-ignored token from the Parser's lexer and updates the Parser's currentToken and peekToken fields
func (p *Parser) next() {
	for {
		// retrieve the next token from the lexer
		peek := p.l.NextToken()
		// if the token is not an ignored token (e.g. whitespace or comments), update the currentToken and peekToken fields and exit the loop
		if !token.IsIgnores(peek.Type) {
			// set the currentToken field to the previous peekToken value
			p.currentToken = p.peekToken
			// set the peekToken field to the new peek value
			p.peekToken = peek
			// exit the loop
			break
		}
	}
}

// currentTokenIs checks if the Parser's currentToken is any of the given token types
func (p *Parser) currentTokenIs(tokens ...token.Type) bool {
	// iterate through the given token types and check if any of them match the currentToken's type
	for _, t := range tokens {
		if p.currentToken.Type == t {
			// if a match is found, return true
			return true
		}
	}
	// if no match is found, return false
	return false
}

// peekTokenIs checks if the Parser's peekToken is any of the given token types
func (p *Parser) peekTokenIs(tokens ...token.Type) bool {
	// iterate through the given token types and check if any of them match the peekToken's type
	for _, t := range tokens {
		if p.peekToken.Type == t {
			// if a match is found, return true
			return true
		}
	}
	// if no match is found, return false
	return false
}

// Error returns an error if there are any errors in the Parser's errors slice
func (p *Parser) Error() error {
	// if there are no errors, return nil
	if len(p.errors) == 0 {
		return nil
	}
	// if there are errors, return the first error message in the errors slice as an error type
	return errors.New(p.errors[0])
}

// Parse reads and parses the input string and returns an AST representation of the schema, along with any errors encountered during parsing
func (p *Parser) Parse() (*ast.Schema, error) {
	// create a new Schema object to store the parsed statements
	schema := &ast.Schema{}
	schema.Statements = []ast.Statement{}

	// loop through the input string until the end is reached
	for !p.currentTokenIs(token.EOF) {
		// parse the next statement in the input string
		stmt, err := p.parseStatement()
		if err != nil {
			// if there was an error parsing the statement, return the error message
			return nil, p.Error()
		}
		if stmt != nil {
			// add the parsed statement to the schema's Statements field if it is not nil
			schema.Statements = append(schema.Statements, stmt)
		}

		// move to the next token in the input string
		p.next()
	}

	// set the schema's references fields to the corresponding maps in the Parser
	schema.SetEntityReferences(p.entityReferences)
	schema.SetRelationReferences(p.relationReferences)
	schema.SetPermissionReferences(p.actionReferences)
	schema.SetRelationalReferences(p.relationalReferences)

	// return the parsed schema object and nil to indicate that there were no errors
	return schema, nil
}

// parseStatement method parses the current statement based on its defined token types
func (p *Parser) parseStatement() (ast.Statement, error) {
	// switch on the currentToken's type to determine which type of statement to parse
	switch p.currentToken.Type {
	case token.ENTITY:
		// if the currentToken is ENTITY, parse an EntityStatement
		return p.parseEntityStatement()
	default:
		// if the currentToken is not recognized, return nil for both the statement and error values
		return nil, nil
	}
}

// parseEntityStatement method parses an ENTITY statement and returns an EntityStatement AST node
func (p *Parser) parseEntityStatement() (*ast.EntityStatement, error) {
	// create a new EntityStatement object and set its Entity field to the currentToken
	stmt := &ast.EntityStatement{Entity: p.currentToken}
	// expect the next token to be an identifier token, and set the EntityStatement's Name field to the identifier's value
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Name = p.currentToken

	// add the entity reference to the Parser's entityReferences map
	err := p.setEntityReference(stmt.Name.Literal)
	if err != nil {
		return nil, err
	}

	// expect the next token to be a left brace token, indicating the start of the entity's body
	if !p.expectAndNext(token.LBRACE) {
		return nil, p.Error()
	}

	// loop through the entity's body until a right brace token is encountered
	for !p.currentTokenIs(token.RBRACE) {
		// if the currentToken is EOF, raise an error and return nil for both the statement and error values
		if p.currentTokenIs(token.EOF) {
			p.currentError(token.RBRACE)
			return nil, p.Error()
		}
		// based on the currentToken's type, parse a RelationStatement or PermissionStatement and add it to the EntityStatement's corresponding field
		switch p.currentToken.Type {
		case token.RELATION:
			relation, err := p.parseRelationStatement(stmt.Name.Literal)
			if err != nil {
				return nil, p.Error()
			}
			stmt.RelationStatements = append(stmt.RelationStatements, relation)
		case token.PERMISSION:
			action, err := p.parsePermissionStatement(stmt.Name.Literal)
			if err != nil {
				return nil, p.Error()
			}
			stmt.PermissionStatements = append(stmt.PermissionStatements, action)
		default:
			// if the currentToken is not recognized, check if it is a newline, left brace, or right brace token, and skip it if it is
			if !p.currentTokenIs(token.NEWLINE) && !p.currentTokenIs(token.LBRACE) && !p.currentTokenIs(token.RBRACE) {
				// if the currentToken is not recognized and not a newline, left brace, or right brace token, raise an error and return nil for both the statement and error values
				p.currentError(token.RELATION, token.PERMISSION)
				return nil, p.Error()
			}
		}
		// move to the next token in the input string
		p.next()
	}

	// return the parsed EntityStatement and nil for the error value
	return stmt, nil
}

// parseRelationStatement method parses a RELATION statement and returns a RelationStatement AST node
func (p *Parser) parseRelationStatement(entityName string) (*ast.RelationStatement, error) {
	// create a new RelationStatement object and set its Relation field to the currentToken
	stmt := &ast.RelationStatement{Relation: p.currentToken}

	// expect the next token to be an identifier token, and set the RelationStatement's Name field to the identifier's value
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Name = p.currentToken
	relationName := stmt.Name.Literal

	// expect the next token to be a SIGN token, indicating the start of the relation type(s)
	if !p.expect(token.SIGN) {
		return nil, p.Error()
	}

	// loop through the relation types until no more SIGN tokens are encountered
	for p.peekTokenIs(token.SIGN) {
		// parse a RelationTypeStatement and append it to the RelationStatement's RelationTypes field
		relationStatement, err := p.parseRelationTypeStatement()
		if err != nil {
			return nil, p.Error()
		}
		stmt.RelationTypes = append(stmt.RelationTypes, *relationStatement)
	}

	// add the relation reference to the Parser's relationReferences and relationalReferences maps
	err := p.setRelationReference(utils.Key(entityName, relationName), stmt.RelationTypes)
	if err != nil {
		return nil, err
	}

	// return the parsed RelationStatement and nil for the error value
	return stmt, nil
}

// parseRelationTypeStatement method parses a single relation type within a RELATION statement and returns a RelationTypeStatement AST node
func (p *Parser) parseRelationTypeStatement() (*ast.RelationTypeStatement, error) {
	// expect the currentToken to be a SIGN token, indicating the start of the relation type
	if !p.expectAndNext(token.SIGN) {
		return nil, p.Error()
	}
	// create a new RelationTypeStatement object and set its Sign field to the SIGN token
	stmt := &ast.RelationTypeStatement{Sign: p.currentToken}

	// expect the next token to be an identifier token, and set the RelationTypeStatement's Type field to the identifier's value
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Type = p.currentToken

	// if the next token is a HASH token, indicating that a specific relation within the relation type is being referenced, parse it and set the RelationTypeStatement's Relation field to the identifier's value
	if p.peekTokenIs(token.HASH) {
		p.next()
		if !p.expectAndNext(token.IDENT) {
			return nil, p.Error()
		}
		stmt.Relation = p.currentToken
	}

	// return the parsed RelationTypeStatement and nil for the error value
	return stmt, nil
}

// parsePermissionStatement method parses an PERMISSION statement and returns an PermissionStatement AST node
func (p *Parser) parsePermissionStatement(entityName string) (ast.Statement, error) {
	// create a new PermissionStatement object and set its Permission field to the currentToken
	stmt := &ast.PermissionStatement{Permission: p.currentToken}

	// expect the next token to be an identifier token, and set the PermissionStatement's Name field to the identifier's value
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Name = p.currentToken

	// add the action reference to the Parser's actionReferences and relationalReferences maps
	err := p.setPermissionReference(utils.Key(entityName, stmt.Name.Literal))
	if err != nil {
		return nil, err
	}

	// expect the next token to be an ASSIGN token, indicating the start of the expression to be assigned to the action
	if !p.expectAndNext(token.ASSIGN) {
		return nil, p.Error()
	}

	p.next()

	// parse the expression statement and set it as the PermissionStatement's ExpressionStatement field
	ex, err := p.parseExpressionStatement()
	if err != nil {
		return nil, p.Error()
	}
	stmt.ExpressionStatement = ex

	// return the parsed PermissionStatement and nil for the error value
	return stmt, nil
}

// parseExpressionStatement method parses an expression statement and returns an ExpressionStatement AST node
func (p *Parser) parseExpressionStatement() (*ast.ExpressionStatement, error) {
	// create a new ExpressionStatement object
	stmt := &ast.ExpressionStatement{}
	var err error
	// parse the expression using the lowest precedence value as the initial precedence level
	stmt.Expression, err = p.parseExpression(LOWEST)
	if err != nil {
		return nil, p.Error()
	}

	// if the next token is a right parenthesis, skip over any additional right parenthesis tokens
	if p.peekTokenIs(token.RPAREN) {
		p.next()
		for p.currentTokenIs(token.RPAREN) {
			p.next()
		}
	}

	// return the parsed ExpressionStatement and nil for the error value
	return stmt, nil
}

// expectAndNext method checks if the next token is of the expected type and advances the lexer to the next token if it is. It returns true if the next token is of the expected type, and false otherwise.
func (p *Parser) expectAndNext(t token.Type) bool {
	// if the next token is of the expected type, advance the lexer to the next token and return true
	if p.peekTokenIs(t) {
		p.next()
		return true
	}
	// otherwise, generate an error message indicating that the expected token type was not found and return false
	p.peekError(t)
	return false
}

// expect method checks if the next token is of the expected type, without advancing the lexer. It returns true if the next token is of the expected type, and false otherwise.
func (p *Parser) expect(t token.Type) bool {
	// if the next token is of the expected type, return true
	if p.peekTokenIs(t) {
		return true
	}
	// otherwise, generate an error message indicating that the expected token type was not found and return false
	p.peekError(t)
	return false
}

// parseExpression method parses an expression with a given precedence level and returns the parsed expression as an AST node. It takes an integer value indicating the precedence level.
func (p *Parser) parseExpression(precedence int) (ast.Expression, error) {
	// if the current token is a left parenthesis, parse the inner expression enclosed in parentheses
	if p.currentTokenIs(token.LPAREN) {
		p.next()
		return p.parseInnerParen()
	}

	// get the prefix parsing function for the current token type
	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil, p.Error()
	}

	// parse the prefix expression
	exp, err := prefix()
	if err != nil {
		return nil, p.Error()
	}

	// continue parsing the expression while the next token has a higher precedence level than the current precedence level
	for !p.peekTokenIs(token.NEWLINE) && precedence < p.peekPrecedence() {
		// get the infix parsing function for the next token type
		infix := p.infixParseFunc[p.peekToken.Type]
		if infix == nil {
			return exp, nil
		}
		p.next()
		// parse the infix expression with the current expression as its left-hand side
		exp, err = infix(exp)
		if err != nil {
			return nil, p.Error()
		}
	}

	// return the parsed expression and nil for the error value
	return exp, nil
}

// parseInnerParen parses the expression inside a pair of parentheses, returning the resulting
// expression and any error encountered.
func (p *Parser) parseInnerParen() (ast.Expression, error) {
	// If the current token is a left parenthesis, parse the expression inside the parentheses.
	if p.currentTokenIs(token.LPAREN) {
		return p.parseExpression(LOWEST)
	}

	// If the current token is not a left parenthesis, it should be the start of an expression.
	// Look up the parsing function for the token type and use it to parse the expression.
	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		// If there is no parsing function for the token type, report an error.
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil, p.Error()
	}
	exp, err := prefix()
	if err != nil {
		return nil, p.Error()
	}

	// Continue parsing the expression until a right parenthesis is encountered.
	for !p.currentTokenIs(token.RPAREN) {
		// If the next token is a right parenthesis, consume it and continue parsing.
		if p.peekTokenIs(token.RPAREN) {
			p.next()
		}
		// Otherwise, the next token should be an infix operator. Look up the parsing function for
		// the token type and use it to parse the expression.
		infix := p.infixParseFunc[p.peekToken.Type]
		if infix == nil {
			// If there is no parsing function for the token type, return the current expression.
			return exp, nil
		}
		p.next()
		exp, err = infix(exp)
		if err != nil {
			return nil, p.Error()
		}
	}

	// If a right parenthesis was encountered, return the parsed expression.
	return exp, nil
}

// parsePrefixExpression parses a prefix expression that starts with an identifier, possibly
// followed by a sequence of dot-separated identifiers, such as "foo.bar.baz".
// It returns the resulting identifier expression and any error encountered.
func (p *Parser) parsePrefixExpression() (ast.Expression, error) {
	// Create a new Identifier expression with the first token as the prefix.
	ident := &ast.Identifier{
		Prefix: p.currentToken,
	}
	// Consume the current token and add it to the list of identifiers.
	p.next()
	ident.Idents = append(ident.Idents, p.currentToken)
	// If the next token is a dot, consume it and continue parsing the next identifier.
	for p.peekTokenIs(token.DOT) {
		p.next()
		p.next()
		ident.Idents = append(ident.Idents, p.currentToken)
	}
	// Return the resulting Identifier expression.
	return ident, nil
}

// parseInfixExpression parses an infix expression that has a left operand and an operator followed by
// a right operand, such as "a or b" or "x and y".
// It takes the left operand as an argument, constructs an InfixExpression with the current operator
// and left operand, and parses the right operand with a higher precedence to construct the final
// expression tree.
// It returns the resulting InfixExpression and any error encountered.
func (p *Parser) parseInfixExpression(left ast.Expression) (ast.Expression, error) {
	// Create a new InfixExpression with the left operand and the current operator.
	expression := &ast.InfixExpression{
		Op:       p.currentToken,
		Left:     left,
		Operator: ast.Operator(p.currentToken.Literal),
	}
	// Get the precedence of the current operator and consume the operator token.
	precedence := p.currentPrecedence()
	p.next()
	// Parse the right operand with a higher precedence to construct the final expression tree.
	ex, err := p.parseExpression(precedence)
	if err != nil {
		return nil, p.Error()
	}
	// Set the right operand of the InfixExpression and return it.
	expression.Right = ex
	return expression, nil
}

// peekPrecedence returns the precedence of the next token in the input, if it is a known
// operator, or the lowest precedence otherwise.
func (p *Parser) peekPrecedence() int {
	if pr, ok := precedences[p.peekToken.Type]; ok {
		return pr
	}
	return LOWEST
}

// currentPrecedence returns the precedence of the current token in the input, if it is a known
// operator, or the lowest precedence otherwise.
func (p *Parser) currentPrecedence() int {
	if pr, ok := precedences[p.currentToken.Type]; ok {
		return pr
	}
	return LOWEST
}

// parseIdentifier parses an identifier expression that may consist of one or more dot-separated
// identifiers, such as "x", "foo.bar", or "a.b.c.d".
// It constructs a new Identifier expression with the first token as the prefix and subsequent
// tokens as identifiers, and returns the resulting expression and any error encountered.
func (p *Parser) parseIdentifier() (ast.Expression, error) {
	// Create a new Identifier expression with the first token as the prefix.
	ident := &ast.Identifier{Idents: []token.Token{p.currentToken}}
	// If the next token is a dot, consume it and continue parsing the next identifier.
	for p.peekTokenIs(token.DOT) {
		p.next()
		p.next()
		ident.Idents = append(ident.Idents, p.currentToken)
	}
	// Return the resulting Identifier expression.
	return ident, nil
}

// registerPrefix registers a parsing function for a prefix token type in the parser's prefixParseFns map.
// It takes a token type and a prefix parsing function as arguments, and stores the function in the map
// under the given token type key.
func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix registers a parsing function for an infix token type in the parser's infixParseFunc map.
// It takes a token type and an infix parsing function as arguments, and stores the function in the map
// under the given token type key.
func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	p.infixParseFunc[tokenType] = fn
}

// duplicationError adds an error message to the parser's error list indicating that a duplication was found.
// It takes a key string as an argument that is used to identify the source of the duplication in the input.
func (p *Parser) duplicationError(key string) {
	msg := fmt.Sprintf("%v:%v:duplication found for %s", p.l.GetLinePosition(), p.l.GetColumnPosition(), key)
	p.errors = append(p.errors, msg)
}

// noPrefixParseFnError adds an error message to the parser's error list indicating that no prefix parsing
// function was found for a given token type.
// It takes a token type as an argument that indicates the type of the token for which a parsing function is missing.
func (p *Parser) noPrefixParseFnError(t token.Type) {
	msg := fmt.Sprintf("%v:%v:no prefix parse function for %s found", p.l.GetLinePosition(), p.l.GetColumnPosition(), t)
	p.errors = append(p.errors, msg)
}

// peekError adds an error message to the parser's error list indicating that the next token in the input
// did not match the expected type(s).
// It takes one or more token types as arguments that indicate the expected types.
func (p *Parser) peekError(t ...token.Type) {
	msg := fmt.Sprintf("%v:%v:expected next token to be %s, got %s instead", p.l.GetLinePosition(), p.l.GetColumnPosition(), t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// currentError adds an error message to the parser's error list indicating that the current token in the input
// did not match the expected type(s).
// It takes one or more token types as arguments that indicate the expected types.
func (p *Parser) currentError(t ...token.Type) {
	msg := fmt.Sprintf("%v:%v:expected token to be %s, got %s instead", p.l.GetLinePosition(), p.l.GetColumnPosition(), t, p.currentToken.Type)
	p.errors = append(p.errors, msg)
}
