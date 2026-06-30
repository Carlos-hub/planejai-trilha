<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Resultado</title>
    <style>
        :root { --accent: #4f46e5; }
        body { font-family: system-ui, sans-serif; max-width: 640px; margin: 0 auto; padding: 2rem 1rem 4rem; line-height: 1.55; color: #1f2937; text-align: center; }
        h1 { font-size: 1.5rem; }
        .pontos { font-size: 2.5rem; font-weight: 800; color: var(--accent); margin: .5rem 0; }
        .bar { background: #e5e7eb; border-radius: 999px; overflow: hidden; height: 26px; margin: 1.5rem 0; }
        .fill { background: #22c55e; height: 100%; text-align: center; color: #fff; font-size: .85rem; line-height: 26px; font-weight: 600; transition: width .4s; }
        a { color: var(--accent); }
    </style>
</head>
<body>
    <h1>Mandou bem, {{ $attempt->nome_aluno }}! 🎉</h1>
    <div class="pontos">{{ $attempt->pontos }} pontos</div>
    <p>{{ $acertos }} de {{ $total }} acertos</p>
    <div class="bar">
        <div class="fill" style="width: {{ $total ? round($acertos / $total * 100) : 0 }}%">
            {{ $total ? round($acertos / $total * 100) : 0 }}%
        </div>
    </div>
    <p><a href="{{ route('trail.show', $trail->codigo) }}">← Voltar à trilha</a></p>
</body>
</html>
