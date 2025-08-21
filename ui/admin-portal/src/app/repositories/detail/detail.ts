import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute } from '@angular/router';
import { MatTableModule } from '@angular/material/table';
import { RepositoriesService, RepositorySummary, ArtifactItem, RepoStats } from '../../core/services/repositories';
import { signal } from '@angular/core';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { DeployDialog } from '../../shared/deploy-dialog/deploy-dialog';
import { MoveDialog } from '../../shared/move-dialog/move-dialog';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { ArtifactsService } from '../../core/services/artifacts';

@Component({
  selector: 'app-detail',
  imports: [CommonModule, MatTableModule, MatDialogModule, MatButtonModule, MatSnackBarModule],
  templateUrl: './detail.html',
  styleUrl: './detail.scss'
})
export class Detail {
  name = '';
  repo = signal<RepositorySummary | null>(null);
  artifacts = signal<ArtifactItem[]>([]);
  stats = signal<RepoStats | null>(null);
  columns = ['path', 'size', 'actions'];

  constructor(private route: ActivatedRoute, private repos: RepositoriesService, private dialog: MatDialog, private snack: MatSnackBar, private artifactsSvc: ArtifactsService) {
    this.name = this.route.snapshot.paramMap.get('name') || '';
    this.load();
  }

  load() {
    if (!this.name) return;
    this.repos.getRepository(this.name).subscribe({ next: r => this.repo.set(r) });
    this.repos.listArtifacts(this.name).subscribe({ next: a => this.artifacts.set(a) });
    this.repos.getStats(this.name).subscribe({ next: s => this.stats.set(s) });
  }

  openDeploy() {
    if (!this.name) return;
    const ref = this.dialog.open(DeployDialog, { data: { repo: this.name } });
    ref.afterClosed().subscribe((res) => {
      if (res?.error) this.snack.open('Deploy failed', 'Dismiss', { duration: 3000 });
      else if (res) this.snack.open('Artifact deployed', undefined, { duration: 2000 });
      this.load();
    });
  }

  openMove(fromPath?: string) {
    if (!this.name) return;
    const ref = this.dialog.open(MoveDialog, { data: { repo: this.name, from: fromPath } });
    ref.afterClosed().subscribe((res) => {
      if (res?.error) this.snack.open('Move failed', 'Dismiss', { duration: 3000 });
      else if (res) this.snack.open('Artifact moved', undefined, { duration: 2000 });
      this.load();
    });
  }

  copy(fromPath: string) {
    if (!this.name) return;
    const toPath = prompt(`Copy from:\n${fromPath}\n\nEnter destination path:`);
    if (!toPath) return;
    this.artifactsSvc.copy(this.name, fromPath, toPath).subscribe({
      next: () => {
        this.snack.open('Artifact copied', undefined, { duration: 2000 });
        this.load();
      },
      error: () => this.snack.open('Copy failed', 'Dismiss', { duration: 3000 })
    });
  }

  downloadUrl(path: string): string {
    // Direct repository download path served by backend; encode path segments but preserve '/'
    const encodedPath = path.split('/').map(encodeURIComponent).join('/');
    return `/api/repositories/${encodeURIComponent(this.name)}/${encodedPath}`;
  }

  delete(path: string) {
    if (!this.name) return;
    const ok = confirm(`Delete artifact ${path}? This cannot be undone.`);
    if (!ok) return;
    this.artifactsSvc.delete(this.name, path).subscribe({
      next: () => {
        this.snack.open('Artifact deleted', undefined, { duration: 2000 });
        this.load();
      },
      error: () => this.snack.open('Delete failed', 'Dismiss', { duration: 3000 })
    });
  }
}
