import pytest

from repo import resolve


def validate_result(graph: resolve.Graph, result: list[str]) -> None:
    assert sorted(graph) == sorted(result), "result should contain exactly vertices from graph"
    for k, deps in graph.items():
        k_idx = result.index(k)
        for dep in deps:
            dep_idx = result.index(dep)
            assert dep_idx < k_idx, \
                f"{k} depends on {dep} but has index ({k_idx}) smaller than its dependency ({dep_idx})\n" \
                f"whole result: {result}"


def test_resolve_graph():
    examples = [
        {
            "0": ["4", "5"],
            "1": ["4", "3"],
            "2": ["5"],
            "3": ["2"],
            "4": [],
            "5": [],
        },
        {
            "m": "",
            "n": "",
            "o": "np",
            "p": "",
            "q": "mn",
            "r": "mos",
            "s": "op",
            "t": "qu",
            "u": "nr",
            "v": "o",
            "w": "v",
            "x": "mv",
            "y": "r",
            "z": "pw",
        },
    ]
    for graph in examples:
        result = resolve.resolve_graph(graph)
        validate_result(graph, result)


def test_cycles():
    examples = [
        {
            "0": ["1"],
            "1": ["2"],
            "2": ["0"],
        },
        {
            "m": "",
            "n": "z",
            "o": "np",
            "p": "",
            "q": "mn",
            "r": "mos",
            "s": "op",
            "t": "qu",
            "u": "nr",
            "v": "o",
            "w": "v",
            "x": "mv",
            "y": "r",
            "z": "pw",
        },
    ]

    for graph in examples:
        with pytest.raises(resolve.CyclicGraphException):
            _ = resolve.resolve_graph(graph)
