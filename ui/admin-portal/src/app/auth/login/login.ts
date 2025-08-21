import { Component } from '@angular/core';
import { AuthService } from '../../core/services/auth';

@Component({
  selector: 'app-login',
  imports: [],
  templateUrl: './login.html',
  styleUrl: './login.scss'
})
export class Login {
  constructor(private auth: AuthService) {}

  signIn() {
    this.auth.login();
  }
}
