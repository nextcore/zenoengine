package slots

import (
	"strings"
	"testing"
)

func TestBladeNativeUnclosedDirectives(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  string
	}{
		{
			name:     "Unclosed @section",
			template: `<html>@section('content')<div>hello</div>`,
			wantErr:  "unclosed @section",
		},
		{
			name:     "Unclosed @if",
			template: `<html>@if(true)<div>hello</div>`,
			wantErr:  "unclosed @if",
		},
		{
			name:     "Unclosed @foreach",
			template: `<html>@foreach($items as $item)<div>hello</div>`,
			wantErr:  "unclosed @foreach",
		},
		{
			name:     "Unclosed @switch",
			template: `<html>@switch($val)`,
			wantErr:  "unclosed @switch",
		},
		{
			name:     "Unclosed @push",
			template: `<html>@push('scripts')<script></script>`,
			wantErr:  "unclosed @push",
		},
		{
			name:     "Unclosed @isset",
			template: `<html>@isset($var)<div>hello</div>`,
			wantErr:  "unclosed @isset",
		},
		{
			name:     "Unclosed @empty",
			template: `<html>@empty($var)<div>hello</div>`,
			wantErr:  "unclosed @empty",
		},
		{
			name:     "Unclosed @unless",
			template: `<html>@unless(true)<div>hello</div>`,
			wantErr:  "unclosed @unless",
		},
		{
			name:     "Unclosed @auth",
			template: `<html>@auth<div>hello</div>`,
			wantErr:  "unclosed @auth",
		},
		{
			name:     "Unclosed @guest",
			template: `<html>@guest<div>hello</div>`,
			wantErr:  "unclosed @guest",
		},
		{
			name:     "Unclosed @for",
			template: `<html>@for($i=0; $i<5; $i++)<div>hello</div>`,
			wantErr:  "unclosed @for",
		},
		{
			name:     "Unclosed @while",
			template: `<html>@while(true)<div>hello</div>`,
			wantErr:  "unclosed @while",
		},
		{
			name:     "Unclosed @forelse",
			template: `<html>@forelse($items as $item)<div>hello</div>`,
			wantErr:  "unclosed @forelse",
		},
		{
			name:     "Unclosed component",
			template: `<html><x-alert><div>hello</div>`,
			wantErr:  "unclosed component x-alert",
		},
		{
			name:     "Unclosed slot",
			template: `<html><x-slot name="header"><div>hello</div>`,
			wantErr:  "unclosed component x-slot",
		},
		{
			name:     "Unclosed @error",
			template: `<html>@error('email')<div>hello</div>`,
			wantErr:  "unclosed @error",
		},
		{
			name:     "Unclosed @can",
			template: `<html>@can('edit')<div>hello</div>`,
			wantErr:  "unclosed @can",
		},
		{
			name:     "Unclosed @cannot",
			template: `<html>@cannot('edit')<div>hello</div>`,
			wantErr:  "unclosed @cannot",
		},
		{
			name:     "Nested Unclosed @section",
			template: `<html>@section('a')<div>@section('b') hello @endsection</div>`,
			wantErr:  "unclosed @section",
		},
		{
			name:     "Nested Unclosed @if",
			template: `<html>@if(true)<div>@if(false) hello @endif</div>`,
			wantErr:  "unclosed @if",
		},
		{
			name:     "Nested Unclosed @foreach",
			template: `<html>@foreach($a as $b)<div>@foreach($c as $d) hello @endforeach</div>`,
			wantErr:  "unclosed @foreach",
		},
		{
			name:     "Nested Unclosed @switch",
			template: `<html>@switch($a)<div>@switch($b) hello @endswitch</div>`,
			wantErr:  "unclosed @switch",
		},
		{
			name:     "Nested Unclosed @push",
			template: `<html>@push('a')<div>@push('b') hello @endpush</div>`,
			wantErr:  "unclosed @push",
		},
		{
			name:     "Nested Unclosed @isset",
			template: `<html>@isset($a)<div>@isset($b) hello @endisset</div>`,
			wantErr:  "unclosed @isset",
		},
		{
			name:     "Nested Unclosed @empty",
			template: `<html>@empty($a)<div>@empty($b) hello @endempty</div>`,
			wantErr:  "unclosed @empty",
		},
		{
			name:     "Nested Unclosed @unless",
			template: `<html>@unless(true)<div>@unless(false) hello @endunless</div>`,
			wantErr:  "unclosed @unless",
		},
		{
			name:     "Nested Unclosed @auth",
			template: `<html>@auth<div>@auth hello @endauth</div>`,
			wantErr:  "unclosed @auth",
		},
		{
			name:     "Nested Unclosed @guest",
			template: `<html>@guest<div>@guest hello @endguest</div>`,
			wantErr:  "unclosed @guest",
		},
		{
			name:     "Nested Unclosed @for",
			template: `<html>@for($i=0;$i<1;$i++)<div>@for($j=0;$j<1;$j++) hello @endfor</div>`,
			wantErr:  "unclosed @for",
		},
		{
			name:     "Nested Unclosed @while",
			template: `<html>@while(true)<div>@while(false) hello @endwhile</div>`,
			wantErr:  "unclosed @while",
		},
		{
			name:     "Nested Unclosed @forelse",
			template: `<html>@forelse($a as $b)<div>@forelse($c as $d) hello @endforelse</div>`,
			wantErr:  "unclosed @forelse",
		},
		{
			name:     "Nested Unclosed component",
			template: `<html><x-alert><div><x-alert> hello </x-alert></div>`,
			wantErr:  "unclosed component x-alert",
		},
		{
			name:     "Nested Unclosed @error",
			template: `<html>@error('a')<div>@error('b') hello @enderror</div>`,
			wantErr:  "unclosed @error",
		},
		{
			name:     "Nested Unclosed @can",
			template: `<html>@can('a')<div>@can('b') hello @endcan</div>`,
			wantErr:  "unclosed @can",
		},
		{
			name:     "Nested Unclosed @cannot",
			template: `<html>@cannot('a')<div>@cannot('b') hello @endcannot</div>`,
			wantErr:  "unclosed @cannot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := transpileBladeNative(tt.template)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}
