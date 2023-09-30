from typing import Mapping, Iterable


"""Graph is represented as a mapping from a name to its dependencies"""
Graph = Mapping[str, Iterable[str]]


class CyclicGraphException(Exception):
    def __init__(self):
        super().__init__("Graph contains cycles")


def resolve_graph(graph: Graph) -> list[str]:
    """Apply topological sorting to the given graph."""
    visited = set()
    resolved = set()
    result = []

    def dfs(v: str):
        visited.add(v)
        for n in graph[v]:
            if n in visited:
                if n not in resolved:
                    raise CyclicGraphException
            else:
                dfs(n)
        resolved.add(v)
        result.append(v)

    for v in graph:
        if v not in visited:
            dfs(v)

    return result
