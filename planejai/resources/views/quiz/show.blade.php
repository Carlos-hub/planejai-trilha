<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Quiz {{ $trail->codigo }}</title>
    <style>
        :root { --accent: #4f46e5; }
        * { box-sizing: border-box; }
        body { font-family: system-ui, sans-serif; max-width: 640px; margin: 0 auto; padding: 2rem 1rem 4rem; line-height: 1.55; color: #1f2937; }
        h1 { font-size: 1.3rem; }
        .nome { display: block; margin: 1rem 0; }
        .nome input { padding: .55rem .7rem; font-size: 1rem; width: 100%; border: 1px solid #d1d5db; border-radius: 8px; }
        .q { border: 1px solid #e5e7eb; border-radius: 10px; padding: 1rem 1.25rem; margin: 1rem 0; }
        .q p.enun { font-weight: 600; margin: 0 0 .6rem; }
        .q label { display: block; padding: .35rem 0; cursor: pointer; }
        button { padding: .75rem 1.3rem; font-size: 1rem; background: var(--accent); color: #fff; border: 0; border-radius: 8px; font-weight: 600; cursor: pointer; }
    </style>
</head>
<body>
    <h1>Quiz — {{ $trail->codigo }}</h1>
    <form method="POST" action="{{ route('quiz.submit', $trail->codigo) }}">
        @csrf
        <label class="nome">Seu nome:
            <input type="text" name="nome_aluno" required>
        </label>
        @foreach ($trail->quiz->questions as $q)
            <div class="q">
                <p class="enun">{{ $loop->iteration }}. {{ $q->enunciado }}</p>
                @foreach ($q->opcoes as $i => $opcao)
                    <label>
                        <input type="radio" name="respostas[{{ $q->id }}]" value="{{ $i }}" required> {{ $opcao }}
                    </label>
                @endforeach
            </div>
        @endforeach
        <button type="submit">Enviar respostas</button>
    </form>
</body>
</html>
