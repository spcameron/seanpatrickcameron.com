package html

import "io"

type Fragment struct {
	Children []Node
}

func (Fragment) isNode() {}

func (f Fragment) Write(w io.Writer) error {
	for _, c := range f.Children {
		if err := c.Write(w); err != nil {
			return err
		}
	}

	return nil
}

func FragmentNode(children ...Node) Fragment {
	if children == nil {
		children = []Node{}
	}

	return Fragment{
		Children: children,
	}
}
