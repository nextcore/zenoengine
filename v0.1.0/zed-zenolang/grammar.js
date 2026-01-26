module.exports = grammar({
  name: 'zenolang',

  rules: {
    source_file: $ => repeat($._definition),

    _definition: $ => choice(
      $.slot_call,
      $.comment
    ),

    comment: $ => token(seq('//', /.*/)),

    slot_call: $ => seq(
      $.slot_name,
      ':',
      optional($._expression),
      optional($.block)
    ),

    slot_name: $ => /[a-zA-Z0-9_\.]+/,

    _expression: $ => choice(
      $.string,
      $.number,
      $.boolean,
      $.null,
      $.variable,
      $.binary_expression
    ),

    block: $ => seq(
      '{',
      repeat($._definition),
      '}'
    ),

    string: $ => choice(
      seq('"', repeat(choice(/[^"\\\n]+|\\\r?\n/, $.escape_sequence)), '"'),
      seq("'", repeat(choice(/[^'\\\n]+|\\\r?\n/, $.escape_sequence)), "'")
    ),

    escape_sequence: $ => token.immediate(seq(
      '\\',
      /(\"|\\|\/|b|f|n|r|t|u[0-9a-fA-F]{4})/
    )),

    number: $ => /\d+(\.\d+)?/,
    boolean: $ => choice('true', 'false'),
    null: $ => choice('nil', 'null'),
    variable: $ => /\$[a-zA-Z0-9_\.]+/,

    keyword_control: $ => choice('if', 'else', 'try', 'catch', 'while', 'include'),
    keyword_other: $ => choice('as', 'do', 'then', 'in'),
    storage_type: $ => choice('var', 'val', 'const'),
    
    operator: $ => choice('+', '-', '*', '/', '%', '==', '!=', '<=', '>=', '<', '>', '='),

    binary_expression: $ => prec.left(1, seq(
        $._expression,
        $.operator,
        $._expression
    ))
  }
});
