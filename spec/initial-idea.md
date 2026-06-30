Aqui vão ideias que atendem os dois lados (professor + aluno) com um core único, o que é estratégico para um MVP solo: você constrói uma engine só e expõe duas interfaces.

# Recomendação principal

**"PlanejAI + Trilha"** — um planejador de aulas alinhado à BNCC (lado professor) que gera, automaticamente, trilhas de estudo focadas para os alunos a partir do mesmo plano. O diferencial competitivo é o vínculo: o professor planeja uma vez, e o aluno recebe conteúdo derivado daquele planejamento, sem trabalho duplicado.

Por que essa ganha nos critérios de avaliação (MVP vale 30%, Inovação 20%, Impacto 20%):

- **Core único, dois produtos**: uma engine de geração via Anthropic API que recebe `{disciplina, série, código BNCC, tempo de aula}` e devolve plano de aula + atividades + trilha do aluno. Isso é viável solo num hackathon.
- **Inovação demonstrável**: o link professor→aluno é o que diferencia de "mais um gerador de plano de aula com IA" (que já existe aos montes). A banca vê novidade real.
- **Impacto mensurável**: reduz horas de planejamento do professor e dá foco ao aluno. Fácil de narrar no pitch com número.

# Funcionalidades do MVP (priorizadas)

Lado professor:
1. Input estruturado: disciplina, ano/série, habilidade BNCC (dropdown ou busca), duração da aula.
2. Geração de plano de aula (objetivos, metodologia, recursos, avaliação) + 1 atividade/quiz.
3. Botão "publicar trilha do aluno" que deriva da aula um roteiro de estudo focado.

Lado aluno:
4. Acesso à trilha (lista de tópicos + resumo + quiz autocorrigido).
5. Modo foco: um tópico por vez, sem dispersão.

Corta tudo o resto. Login pode ser fake/simplificado no MVP — a banca aceita isso desde que você explique (está nas regras do PDF).

# Stack sugerida (alinhada ao que você domina)

Laravel monólito modular, dois módulos (`Planning`, `Learning`) compartilhando a `AiService` que encapsula a Anthropic API. PostgreSQL. Blade + Alpine ou Livewire pro front rápido — evita gastar tempo de hackathon montando SPA. A BNCC você carrega como seed (JSON estático das habilidades por ano), não precisa de fonte externa em tempo real.

# Alternativas (se quiser fugir do óbvio)

**Diferenciador de inclusão digital (critério forte no PDF):** versão da trilha do aluno que funciona em baixíssima banda e exporta pra compartilhar via WhatsApp (texto/PDF leve). Isso ataca o critério de Inclusão Digital que a maioria das equipes ignora — é onde você ganha pontos de Inovação com pouco esforço técnico.

**Engajamento via gamificação:** quiz da trilha com pontuação e progresso. Barato de implementar (contador + barra), alto retorno na demo do MVP.

# Risco a sinalizar

O maior gargalo de um projeto solo com IA num hackathon não é técnico, é **escopo**. A tentação de fazer professor + aluno + gamificação + WhatsApp vai te afundar. Escolha o eixo professor→aluno como espinha dorsal e trate o resto como "próximos passos" no pitch (que é uma seção obrigatória da entrega, então você ganha pontos falando do que *não* fez).

Quer que eu detalhe a arquitetura modular Laravel desses dois módulos, ou prefere começar pela seed da BNCC e o contrato da `AiService`?