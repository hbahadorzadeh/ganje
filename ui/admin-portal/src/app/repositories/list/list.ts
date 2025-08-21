import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { Router, RouterLink } from '@angular/router';
import { RepositoriesService, RepositorySummary } from '../../core/services/repositories';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';

@Component({
  selector: 'app-list',
  imports: [CommonModule, MatTableModule, MatButtonModule, MatIconModule, RouterLink, MatSnackBarModule],
  templateUrl: './list.html',
  styleUrl: './list.scss'
})
export class List {
  displayedColumns = ['name', 'artifactType', 'actions'];
  readonly data = signal<RepositorySummary[]>([]);

  constructor(private repos: RepositoriesService, private router: Router, private snack: MatSnackBar) {
    this.load();
  }

  load() {
    this.repos.listRepositories().subscribe({
      next: (items) => {
        this.data.set(items);
        this.snack.open(`${items.length} repos`, undefined, { duration: 1500 });
      },
      error: () => {
        this.data.set([]);
        this.snack.open('Failed to load repositories', 'Dismiss', { duration: 3000 });
      },
    });
  }

  open(repo: RepositorySummary) {
    this.router.navigate(['/repositories', repo.name]);
  }
}
