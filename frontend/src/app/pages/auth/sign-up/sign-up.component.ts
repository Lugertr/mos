import { Component, DestroyRef, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { AuthService } from '../../../core/auth/auth.service';
import { LoadingBarService } from '@core/loading-bar/loading-bar.service';
import { InformerService } from '@core/services/informer.service';
import { takeUntilDestroyed, toSignal } from '@angular/core/rxjs-interop';
import { PathWithSlash, RoutesPath } from 'src/app/app.routes';

@Component({
  standalone: true,
  selector: 'app-sign-up',
  imports: [
    ReactiveFormsModule,
    MatFormFieldModule, MatInputModule, MatButtonModule, MatCardModule, MatSnackBarModule,RouterModule
  ],
  templateUrl: './sign-up.component.html',
  styleUrls: ['../auth.scss'],
})
export class SignUpComponent {
  private readonly fb = inject(FormBuilder);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);
  private readonly informerSrv = inject(InformerService);
  private readonly loadingBarSrv = inject(LoadingBarService);

  readonly isLoading = toSignal(this.loadingBarSrv.show$);
  form = this.fb.group({
    login: ['', [Validators.required]],
    password: ['', [Validators.required, Validators.minLength(6)]],
    full_name: ['', [Validators.required]],
  });

  get signInLink(): string {
    return PathWithSlash(RoutesPath.SignIn);
  }

  onSubmit(): void {
    if (this.form.invalid) return;

    const payload = {
      login: this.form.value.login!,
      password: this.form.value.password!,
      full_name: this.form.value.full_name,
    };

    this.auth.signUp(payload).pipe(
      this.loadingBarSrv.withLoading(),
      takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (res) => {
        if (res?.token) this.auth.saveToken(res.token);
        this.router.navigateByUrl('/');
      },
        error: (err) => this.informerSrv.error(err?.error?.message, 'Ошибка входа'),
    });
  }
}
