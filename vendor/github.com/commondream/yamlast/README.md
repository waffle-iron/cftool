# yaml-ast

This is a horrible copy and paste job of the underlying yaml parser in
[go-yaml](https://github.com/go-yaml/yaml) because I wanted to be able to get
an AST out of the yaml parser, instead of just marshalling and unmarshalling
yaml into structures. Maybe you'll find it useful?

## Installing

```
go get -u github.com/commondream/yaml-ast
```

## Usage

You'll mainly be interested in using the `Parse` function to generate a `Node`
pointer to the document node. Then you can iterate nodes.

```
node := Parse(data)
for _, child in range node.children {
  fmt.Println(child.Value)
}
```

## License

See the LICENSE file for more details on the licensing of this project. Various
files in the project are licensed differently, based on their origin project.
