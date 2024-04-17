//
//  bodyText.swift
//  notes
//
//  Created by Cyril Flambard on 17/04/2024.
//

import SwiftUI

struct bodyText: View {
        @State private var content: String
        @State private var isShowingTitle = true
        @FocusState private var isFocused : Bool
        
        init(content: String = "Content") {
            self._content = State(initialValue: content)
        }
        
        var body: some View {
            HStack() {
                if isShowingTitle {
                    Text(content)
                        .multilineTextAlignment(.leading)
                        .onTapGesture {
                            isShowingTitle.toggle()
                        }
                } else {
                    
                    TextEditor(text: $content)
                        .frame(minWidth: 200, minHeight: 100) // Taille minimale
                        .onSubmit() {
                            isShowingTitle.toggle()
                        }
                        .focused($isFocused)
                        .onChange(of: isFocused) { () -> Void in
                            if isFocused {
                                print("focus body")
                            } else {
                                print("unfocus body")
                                isShowingTitle.toggle()
                            }
                        }
                }
                Spacer()
            }
        }
    
        /*
         
            Si une structure est une class, elle doit y avoir des méthodes ?
            Alors je devrais lui faire des méthodes de style.
         
         */
    
    }

#Preview {
    bodyText(content: "Le contenue de mon article")
}
