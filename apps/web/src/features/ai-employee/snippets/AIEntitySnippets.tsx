import {
  entitySnippetStatus,
  entitySnippetTitle,
  textValue,
} from "../aiEmployeeUtils";

export function AIEntitySnippets({ snippets }: { snippets: Record<string, unknown>[] }) {
  if (!snippets.length) return null;
  return (
    <div className="ai-entity-snippets" data-testid="ai-entity-snippets">
      {snippets.map((snippet, index) => {
        const selectedURL = textValue(snippet.selected_url);
        return (
          <div className="ai-entity-snippet" key={`${textValue(snippet.entity_type)}-${textValue(snippet.query)}-${index}`}>
            <span>{entitySnippetTitle(snippet)} · {textValue(snippet.query) || "без запроса"}</span>
            {selectedURL ? <a href={selectedURL}>{entitySnippetStatus(snippet)}</a> : <strong>{entitySnippetStatus(snippet)}</strong>}
            <em>{textValue(snippet.tool_key) || "entity.search"}</em>
          </div>
        );
      })}
    </div>
  );
}
