import { Component} from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { NavbarComponent } from './components/navbar/navbar.component';
import { FooterComponent } from './components/footer/footer.component';
import { MenuComponent } from './components/menu/menu.component';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { LoadingBarComponent } from '@core/loading-bar/loading-bar.component';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss',
  imports: [
    RouterOutlet,
    NavbarComponent,
    FooterComponent,
    MenuComponent,
    LoadingBarComponent
  ],
})
export class AppComponent {
}
