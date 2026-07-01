<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class CurriculumUnit extends Model
{
    protected $fillable = ['etapa', 'serie', 'componente', 'origem', 'trimestre', 'objetos'];

    protected $casts = ['objetos' => 'array'];
}
