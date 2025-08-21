import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators, FormGroup } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { Inject } from '@angular/core';
import { ArtifactsService } from '../../core/services/artifacts';

@Component({
  selector: 'app-move-dialog',
  imports: [CommonModule, ReactiveFormsModule, MatFormFieldModule, MatInputModule, MatButtonModule, MatDialogModule],
  templateUrl: './move-dialog.html',
  styleUrl: './move-dialog.scss'
})
export class MoveDialog {
  form!: FormGroup;

  constructor(
    private fb: FormBuilder,
    private artifacts: ArtifactsService,
    private dialogRef: MatDialogRef<MoveDialog>,
    @Inject(MAT_DIALOG_DATA) public data: { repo: string; from?: string }
  ) {
    this.form = this.fb.group({
      from: [data?.from || '', Validators.required],
      to: ['', Validators.required],
    });
  }

  submit() {
    if (this.form.invalid) return;
    const { from, to } = this.form.value as { from: string; to: string };
    this.artifacts.move(this.data.repo, from, to).subscribe({
      next: (res) => this.dialogRef.close(res),
      error: (err) => this.dialogRef.close({ error: err })
    });
  }

  cancel() {
    this.dialogRef.close();
  }
}
