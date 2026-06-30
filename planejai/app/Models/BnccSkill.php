<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class BnccSkill extends Model
{
    protected $fillable = ['code', 'disciplina', 'ano', 'descricao'];
}
