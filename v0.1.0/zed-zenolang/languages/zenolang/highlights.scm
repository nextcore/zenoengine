;; Keywords
(keyword_control) @keyword.control
(keyword_other) @keyword
(storage_type) @type
(boolean) @constant.builtin
(null) @constant.builtin

;; Strings
(string) @string

;; Numbers
(number) @constant.numeric

;; Comments
(comment) @comment

;; Slots (Built-in)
(slot_call
  (slot_name) @function.builtin
  (#match? @function.builtin "^(http|db|view|sec|io|os|strings|math|validator|log|template|array|map|loop|crypto|job|image|mail|monitor|system|cast|validate)\\.")
)

;; Slots (Custom)
(slot_call
  (slot_name) @function
)

;; Variables
(variable) @variable

;; Operators
(operator) @operator
