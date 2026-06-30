<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: DejaVu Sans, sans-serif; font-size: 12px; color: #222; }
        h1 { font-size: 16px; border-bottom: 2px solid #4f46e5; padding-bottom: 6px; }
        .t { margin: 10px 0; padding: 8px 10px; border: 1px solid #ddd; border-radius: 6px; }
        .t strong { color: #4f46e5; }
        .codigo { color: #4f46e5; font-weight: bold; }
    </style>
</head>
<body>
    <h1>Trilha <span class="codigo">{{ $trail->codigo }}</span> — {{ $trail->lessonPlan->bnccSkill->disciplina }}</h1>
    <p>{{ $trail->lessonPlan->bnccSkill->ano }} · {{ $trail->lessonPlan->bnccSkill->code }}</p>
    @foreach ($trail->topics as $topico)
        <div class="t"><strong>{{ $topico->ordem }}. {{ $topico->titulo }}</strong><br>{{ $topico->resumo }}</div>
    @endforeach
</body>
</html>
