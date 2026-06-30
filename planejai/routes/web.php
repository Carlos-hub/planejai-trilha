<?php

use App\Http\Controllers\ExportController;
use App\Http\Controllers\QuizController;
use App\Http\Controllers\TrailController;
use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    return view('welcome');
});

Route::get('/t/{codigo}', [TrailController::class, 'show'])->name('trail.show');
Route::get('/t/{codigo}/quiz', [QuizController::class, 'show'])->name('quiz.show');
Route::post('/t/{codigo}/quiz', [QuizController::class, 'submit'])->name('quiz.submit');
Route::get('/t/{codigo}/pdf', [ExportController::class, 'pdf'])->name('trail.pdf');
