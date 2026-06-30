<x-filament-panels::page>
    @php($stats = $this->getStats())
    @php($attempts = $this->getAttempts())

    <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <x-filament::section>
            <x-slot name="heading">Alunos</x-slot>
            <div class="text-3xl font-bold">{{ $stats['total_alunos'] }}</div>
        </x-filament::section>
        <x-filament::section>
            <x-slot name="heading">Média de pontos</x-slot>
            <div class="text-3xl font-bold">{{ $stats['media_pontos'] }}</div>
        </x-filament::section>
        <x-filament::section>
            <x-slot name="heading">Concluíram</x-slot>
            <div class="text-3xl font-bold">{{ $stats['concluidos'] }}</div>
        </x-filament::section>
    </div>

    <x-filament::section>
        <x-slot name="heading">Tentativas</x-slot>

        @if ($attempts->isEmpty())
            <p class="text-sm text-gray-500">Nenhum aluno fez o quiz ainda.</p>
        @else
            <table class="w-full text-sm">
                <thead>
                    <tr class="text-left border-b">
                        <th class="py-2">Aluno</th>
                        <th class="py-2">Pontos</th>
                        <th class="py-2">Concluído em</th>
                    </tr>
                </thead>
                <tbody>
                    @foreach ($attempts as $a)
                        <tr class="border-b border-gray-100">
                            <td class="py-2">{{ $a->nome_aluno }}</td>
                            <td class="py-2">{{ $a->pontos }}</td>
                            <td class="py-2">{{ $a->concluido_em?->format('d/m/Y H:i') ?? '—' }}</td>
                        </tr>
                    @endforeach
                </tbody>
            </table>
        @endif
    </x-filament::section>
</x-filament-panels::page>
