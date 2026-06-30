<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Trilha {{ $trail->codigo }}</title>
    <style>
        :root { --accent: #4f46e5; }
        * { box-sizing: border-box; }
        body { font-family: system-ui, sans-serif; max-width: 640px; margin: 0 auto; padding: 2rem 1rem 4rem; line-height: 1.55; color: #1f2937; }
        header { border-bottom: 2px solid var(--accent); padding-bottom: 1rem; margin-bottom: 1.5rem; }
        h1 { font-size: 1.4rem; margin: 0 0 .25rem; }
        .codigo { display: inline-block; background: #eef2ff; color: var(--accent); font-weight: 600; padding: .15rem .5rem; border-radius: 6px; letter-spacing: .05em; }
        .topico { border: 1px solid #e5e7eb; border-radius: 10px; padding: 1rem 1.25rem; margin: 1rem 0; box-shadow: 0 1px 2px rgba(0,0,0,.04); }
        .topico h2 { font-size: 1.05rem; margin: 0 0 .35rem; }
        .topico p { margin: 0; color: #4b5563; }
        .cta { display: inline-block; background: var(--accent); color: #fff; text-decoration: none; padding: .7rem 1.2rem; border-radius: 8px; font-weight: 600; margin-top: .5rem; }
        .share { margin-top: 1.5rem; font-size: .9rem; }
        .share a { color: var(--accent); }
    </style>
</head>
<body>
    <header>
        <h1>{{ $trail->lessonPlan->bnccSkill->disciplina }} — {{ $trail->lessonPlan->bnccSkill->ano }}</h1>
        <p>Código da trilha: <span class="codigo">{{ $trail->codigo }}</span></p>
    </header>

    @foreach ($trail->topics as $topico)
        <div class="topico">
            <h2>{{ $topico->ordem }}. {{ $topico->titulo }}</h2>
            <p>{{ $topico->resumo }}</p>
        </div>
    @endforeach

    @if ($trail->quiz)
        <p><a class="cta" href="{{ route('quiz.show', $trail->codigo) }}">Fazer o quiz →</a></p>
    @endif

    <p class="share">
        <a href="{{ route('trail.pdf', $trail->codigo) }}">Baixar PDF</a>
        &nbsp;|&nbsp;
        <a href="https://wa.me/?text={{ urlencode('Estude esta trilha: ' . route('trail.show', $trail->codigo)) }}" target="_blank" rel="noopener">Compartilhar no WhatsApp</a>
    </p>
</body>
</html>
