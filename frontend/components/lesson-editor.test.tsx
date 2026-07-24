import { describe, it, expect, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";

// LessonMeta fetches the BNCC catalog on mount and is irrelevant to the Plano
// de aula flow under test — stub it so tests stay focused and offline.
vi.mock("@/components/lesson-meta", () => ({
  LessonMeta: () => <div data-testid="lesson-meta" />,
}));

import {
  LessonEditor,
  EMPTY_LESSON_CONTENT,
  PLANO_STEPS,
  type LessonContent,
} from "@/components/lesson-editor";

const clone = (c: LessonContent): LessonContent => JSON.parse(JSON.stringify(c));
const step = (id: string) => PLANO_STEPS.find((s) => s.id === id)!;
const li = (id: string) =>
  document.querySelector(`[data-step="${id}"]`) as HTMLElement;
const textarea = (id: string) =>
  within(li(id)).getByPlaceholderText(step(id).placeholder) as HTMLTextAreaElement;
const exampleButton = (id: string) =>
  within(li(id)).getByText(step(id).example).closest("button") as HTMLButtonElement;

/** Controlled wrapper so the editor behaves like it does in the app. */
function Harness() {
  const [content, setContent] = useState<LessonContent>(() =>
    clone(EMPTY_LESSON_CONTENT)
  );
  return (
    <LessonEditor
      value={content}
      onChange={setContent}
      bnccSkillId={null}
      onBnccSkillIdChange={() => {}}
      duracao={50}
      onDuracaoChange={() => {}}
      onSave={() => {}}
    />
  );
}

describe("LessonEditor — Plano de aula sequencial", () => {
  it("renders all five steps, in order, each numbered", () => {
    render(<Harness />);
    const ids = [...document.querySelectorAll("[data-step]")].map((el) =>
      el.getAttribute("data-step")
    );
    expect(ids).toEqual([
      "objetivos",
      "metodologia",
      "recursos",
      "avaliacao",
      "atividade",
    ]);
    for (const s of PLANO_STEPS) {
      expect(
        within(li(s.id)).getByRole("heading", { name: s.label })
      ).toBeInTheDocument();
    }
    // Unfilled steps show their step number in the circle.
    expect(within(li("objetivos")).getByText("1")).toBeInTheDocument();
    expect(within(li("atividade")).getByText("5")).toBeInTheDocument();
  });

  it("starts with zero progress", () => {
    render(<Harness />);
    expect(screen.getByText("0 de 5 etapas")).toBeInTheDocument();
  });

  it("clicking an example seeds only that empty field and advances progress", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.click(exampleButton("objetivos"));

    expect(textarea("objetivos").value).toBe(step("objetivos").example);
    expect(li("objetivos").getAttribute("data-filled")).toBe("true");
    expect(screen.getByText("1 de 5 etapas")).toBeInTheDocument();
    // Siblings remain untouched.
    expect(textarea("metodologia").value).toBe("");
    expect(li("metodologia").getAttribute("data-filled")).toBe("false");
  });

  it("does not overwrite text the teacher already wrote", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(textarea("recursos"), "Projetor e livro");
    expect(exampleButton("recursos")).toBeDisabled();
    // A disabled button can't fire, so the text survives.
    expect(textarea("recursos").value).toBe("Projetor e livro");
  });

  it("typing marks the step filled and updates the progress count", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.type(textarea("avaliacao"), "Quiz final");
    expect(li("avaliacao").getAttribute("data-filled")).toBe("true");
    expect(screen.getByText("1 de 5 etapas")).toBeInTheDocument();
  });
});

describe("LessonEditor — field routing", () => {
  it("writes plano steps into content.plano and Atividade into content.atividade", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <LessonEditor
        value={clone(EMPTY_LESSON_CONTENT)}
        onChange={onChange}
        bnccSkillId={null}
        onBnccSkillIdChange={() => {}}
        duracao={50}
        onDuracaoChange={() => {}}
        onSave={() => {}}
      />
    );

    await user.click(exampleButton("objetivos"));
    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({
        plano: expect.objectContaining({ objetivos: step("objetivos").example }),
      })
    );

    onChange.mockClear();
    await user.click(exampleButton("atividade"));
    const arg = onChange.mock.calls[0][0] as LessonContent;
    // Atividade lives at the top level, not inside plano.
    expect(arg.atividade).toBe(step("atividade").example);
    expect(arg.plano.objetivos).toBe("");
  });
});
