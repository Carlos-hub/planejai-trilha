import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";

export type LessonMode = "manual" | "ia" | "enhance";

interface ModeOption {
  mode: LessonMode;
  title: string;
  description: string;
  action: string;
}

const MODE_OPTIONS: ModeOption[] = [
  {
    mode: "manual",
    title: "Criar manualmente",
    description: "Monte o plano de aula, tópicos e quiz do zero, com controle total sobre cada campo.",
    action: "Criar manualmente",
  },
  {
    mode: "ia",
    title: "Gerar com IA",
    description: "Escolha a habilidade BNCC e a duração; a IA gera o plano completo, a trilha e o quiz.",
    action: "Gerar com IA",
  },
  {
    mode: "enhance",
    title: "Aprimorar rascunho com IA",
    description: "Escreva um rascunho inicial e deixe a IA revisar e enriquecer o conteúdo.",
    action: "Aprimorar com IA",
  },
];

export function ModePicker({ onPick }: { onPick: (mode: LessonMode) => void }) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
      {MODE_OPTIONS.map((option) => (
        <Card key={option.mode}>
          <CardHeader>
            <CardTitle>{option.title}</CardTitle>
            <CardDescription>{option.description}</CardDescription>
          </CardHeader>
          <CardContent />
          <CardFooter>
            <Button className="w-full" onClick={() => onPick(option.mode)}>
              {option.action}
            </Button>
          </CardFooter>
        </Card>
      ))}
    </div>
  );
}
