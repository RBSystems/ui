import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';

import {
  MatButtonModule,
  MatFormFieldModule,
  MatInputModule,
  MatIconModule,
  MatButtonToggleModule,
  MatBottomSheetModule,
  MatToolbarModule,
  MatSliderModule,
  MatDialogModule,
  MatTabsModule,
  MatDividerModule
} from '@angular/material';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { BFFService } from './services/bff.service';
import { LoginComponent } from './components/login/login.component';
import { NumpadComponent } from './dialogs/numpad/numpad.component';
import { SquareButtonComponent } from './components/square-button/square-button.component';
import { VolumeSliderComponent } from './components/volume-slider/volume-slider.component';
import { DisplayDialogComponent } from './dialogs/display-dialog/display-dialog.component';
import { RoomControlComponent } from './components/room-control/room-control.component';
import { SelectionComponent } from './components/selection/selection.component';
import { AudioComponent } from './components/audio/audio.component';
import { PresentComponent } from './components/present/present.component';
import { HelpComponent } from './components/help/help.component';
import { SingleDisplayComponent } from './components/single-display/single-display.component';
import { MultiDisplayComponent } from './components/multi-display/multi-display.component';
import { HelpInfoComponent } from './components/help/help-info/help-info.component';
import { WideButtonComponent } from './components/wide-button/wide-button.component';
import { ControlTabComponent } from './components/control-tab/control-tab.component';
import { ControlTabDirective } from './components/control-tab/control-tab.directive';

@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    NumpadComponent,
    SquareButtonComponent,
    VolumeSliderComponent,
    DisplayDialogComponent,
    RoomControlComponent,
    SelectionComponent,
    AudioComponent,
    PresentComponent,
    HelpComponent,
    SingleDisplayComponent,
    MultiDisplayComponent,
    HelpInfoComponent,
    WideButtonComponent,
    ControlTabComponent,
    ControlTabDirective
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    AppRoutingModule,
    MatButtonModule,
    MatFormFieldModule,
    FormsModule,
    MatInputModule,
    MatIconModule,
    MatButtonToggleModule,
    MatToolbarModule,
    MatSliderModule,
    MatDialogModule,
    ReactiveFormsModule,
    MatBottomSheetModule,
    HttpClientModule,
    MatTabsModule,
    MatDividerModule,
  ],
  providers: [
    BFFService
  ],
  entryComponents: [
    NumpadComponent,
    DisplayDialogComponent,
    HelpInfoComponent,
    SingleDisplayComponent,
    MultiDisplayComponent,
    AudioComponent,
    PresentComponent,
    HelpComponent
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }