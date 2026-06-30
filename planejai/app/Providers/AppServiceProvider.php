<?php

namespace App\Providers;

use App\Contracts\LessonGenerator;
use App\Services\PrismLessonGenerator;
use Illuminate\Support\ServiceProvider;

class AppServiceProvider extends ServiceProvider
{
    /**
     * Register any application services.
     */
    public function register(): void
    {
        $this->app->bind(LessonGenerator::class, function ($app) {
            return new PrismLessonGenerator(
                provider: config('llm.provider'),
                model: config('llm.model'),
            );
        });
    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        //
    }
}
