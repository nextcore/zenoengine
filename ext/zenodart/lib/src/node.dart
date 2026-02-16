class Node {
  String name;
  dynamic value; // String, usually
  List<Node> children;
  Node? parent;
  int line;
  int col;
  String filename;

  Node({
    this.name = "",
    this.value,
    List<Node>? children,
    this.parent,
    this.line = 0,
    this.col = 0,
    this.filename = "",
  }) : children = children ?? [];

  @override
  String toString() {
    return 'Node(name: $name, value: "$value", children: ${children.length}, line: $line)';
  }

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      'value': value,
      'children': children.map((c) => c.toJson()).toList(),
      'line': line,
      'col': col,
    };
  }
}
