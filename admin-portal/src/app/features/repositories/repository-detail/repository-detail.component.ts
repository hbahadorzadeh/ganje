import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';

@Component({
  selector: 'app-repository-detail',
  standalone: true,
  imports: [CommonModule, MatCardModule, MatButtonModule],
  template: `
    <mat-card>
      <mat-card-header>
        <mat-card-title>Repository Details</mat-card-title>
      </mat-card-header>
      <mat-card-content>
        <p>Repository details will be displayed here.</p>
        <button mat-raised-button color="primary">Edit Repository</button>
      </mat-card-content>
    </mat-card>
  `,
  styles: [`
    .container { padding: 2rem; max-width: 800px; margin: 0 auto; }
  `]
})
export class RepositoryDetailComponent {}
