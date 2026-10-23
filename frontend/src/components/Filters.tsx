import type { CategoryNode, Tag } from '../lib/types'

export function CategoryTree({
  nodes,
  selected,
  onSelect,
  depth = 0,
}: {
  nodes: CategoryNode[]
  selected: string
  onSelect: (slug: string) => void
  depth?: number
}) {
  return (
    <ul className={depth > 0 ? 'ml-3 border-l-[1.5px] border-mist pl-3' : ''}>
      {nodes.map((node) => (
        <li key={node.id}>
          <button
            type="button"
            onClick={() => onSelect(node.slug)}
            aria-current={selected === node.slug ? 'true' : undefined}
            className={`w-full rounded-md px-2 py-1.5 text-left transition-colors hover:bg-mist/30 ${
              selected === node.slug ? 'bg-ink font-semibold text-paper hover:bg-ink' : ''
            }`}
          >
            {node.name}
          </button>
          {node.children.length > 0 && (
            <CategoryTree nodes={node.children} selected={selected} onSelect={onSelect} depth={depth + 1} />
          )}
        </li>
      ))}
    </ul>
  )
}

export function TagList({ tags, selected, onSelect }: { tags: Tag[]; selected: string; onSelect: (slug: string) => void }) {
  return (
    <div className="flex flex-wrap gap-2">
      {tags.map((tag) => (
        <button
          key={tag.id}
          type="button"
          className="chip"
          aria-pressed={selected === tag.slug}
          onClick={() => onSelect(tag.slug)}
        >
          {tag.name}
        </button>
      ))}
    </div>
  )
}

export function findCategory(nodes: CategoryNode[], slug: string): CategoryNode | undefined {
  for (const node of nodes) {
    if (node.slug === slug) return node
    const inChildren = findCategory(node.children, slug)
    if (inChildren) return inChildren
  }
  return undefined
}
