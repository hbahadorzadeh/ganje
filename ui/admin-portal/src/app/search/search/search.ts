import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatTableModule } from '@angular/material/table';
import { SearchService, SearchResultItem } from '../../core/services/search';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';

@Component({
  selector: 'app-search',
  imports: [CommonModule, FormsModule, MatTableModule, MatSnackBarModule],
  templateUrl: './search.html',
  styleUrl: './search.scss'
})
export class Search {
  q = '';
  repository = '';
  columns = ['repository', 'path', 'size'];
  readonly results = signal<SearchResultItem[]>([]);

  constructor(private searchSvc: SearchService, private snack: MatSnackBar) {}

  run() {
    if (!this.q) { this.results.set([]); return; }
    this.searchSvc.search(this.q, this.repository || undefined)
      .subscribe({
        next: (r) => {
          this.results.set(r);
          this.snack.open(`${r.length} result(s)`, undefined, { duration: 1500 });
        },
        error: () => {
          this.results.set([]);
          this.snack.open('Search failed', 'Dismiss', { duration: 3000 });
        }
      });
  }
}
