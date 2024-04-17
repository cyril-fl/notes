//
//  bodyText.swift
//  notes
//
//  Created by Cyril Flambard on 17/04/2024.
//

import SwiftUI

struct titleText: View {
    @State private var title: String
    @State private var isShowingTitle = true
    @FocusState private var isFocused : Bool
    
    init(title: String = "Titre") {
        self._title = State(initialValue: title)
    }
    
    var body: some View {
        HStack() {
            if isShowingTitle {
                Text(title)
                    .onTapGesture {
                        isShowingTitle.toggle()
                    }
                    
            } else {
                TextField("Entrez un titre", text: $title)
                    .onSubmit() {
                        isShowingTitle.toggle()
                    }
                    .focused($isFocused)
                    .onChange(of: isFocused) { () -> Void in
                        if isFocused {
                            print("focus Title")
                        } else {
                            print("unfocus Title")
                            isShowingTitle.toggle()
                        }
                    }
            }

        }
    }
    
    /*
     
        Si une structure est une class, elle doit y avoir des méthodes ?
        Alors je devrais lui faire des méthodes de style.
     
     */
}

#Preview {
    titleText(title: "Mon titre")
}

